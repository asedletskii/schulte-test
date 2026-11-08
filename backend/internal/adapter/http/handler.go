package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	nethttp "net/http"

	"schulte/internal/domain"
	"schulte/internal/usecase"
)

//содержит прикладные сервисы, вызываемые из HTTP-обработчиков
type Handler struct {
	auth     *usecase.AuthService
	subjects *usecase.SubjectService
	tests    *usecase.TestService
	results  *usecase.ResultService
	settings *usecase.SettingsService
	log      *slog.Logger
}

// зависимости HTTP-слоя
type Deps struct {
	Auth     *usecase.AuthService
	Subjects *usecase.SubjectService
	Tests    *usecase.TestService
	Results  *usecase.ResultService
	Settings *usecase.SettingsService
	Log      *slog.Logger
}

func NewHandler(d Deps) *Handler {
	return &Handler{auth: d.Auth, subjects: d.Subjects, tests: d.Tests, results: d.Results, settings: d.Settings, log: d.Log}
}

func (h *Handler) writeJSON(w nethttp.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

type errorBody struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func (h *Handler) writeError(w nethttp.ResponseWriter, err error) {
	status, code := nethttp.StatusInternalServerError, "internal"
	msg := "внутренняя ошибка сервера"
	switch {
	case errors.Is(err, domain.ErrValidation):
		status, code, msg = nethttp.StatusBadRequest, "validation", err.Error()
	case errors.Is(err, domain.ErrUnauthorized):
		status, code, msg = nethttp.StatusUnauthorized, "unauthorized", err.Error()
	case errors.Is(err, domain.ErrForbidden):
		status, code, msg = nethttp.StatusForbidden, "forbidden", err.Error()
	case errors.Is(err, domain.ErrNotFound):
		status, code, msg = nethttp.StatusNotFound, "not_found", err.Error()
	case errors.Is(err, domain.ErrConflict):
		status, code, msg = nethttp.StatusConflict, "conflict", err.Error()
	default:
		h.log.Error("внутренняя ошибка", "err", err)
	}
	h.writeJSON(w, status, errorBody{Error: msg, Code: code})
}

func decode(r *nethttp.Request, dst any) error {
	r.Body = nethttp.MaxBytesReader(nil, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return domain.Validationf("некорректное тело запроса: %v", err)
	}
	return nil
}

// ---- аутентификация ----

func (h *Handler) loginAdmin(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req adminLoginRequest
	if err := decode(r, &req); err != nil {
		h.writeError(w, err)
		return
	}
	s, err := h.auth.LoginAdmin(r.Context(), req.Password)
	h.respondSession(w, s, err)
}

func (h *Handler) loginSubject(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req subjectLoginRequest
	if err := decode(r, &req); err != nil {
		h.writeError(w, err)
		return
	}
	s, err := h.auth.LoginSubject(r.Context(), req.SubjectID)
	h.respondSession(w, s, err)
}

func (h *Handler) respondSession(w nethttp.ResponseWriter, s usecase.Session, err error) {
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeJSON(w, nethttp.StatusOK, sessionResponse{
		Token: s.Token, ExpiresAt: s.ExpiresAt, Role: string(s.Principal.Role), SubjectID: s.Principal.SubjectID,
	})
}

func (h *Handler) me(w nethttp.ResponseWriter, r *nethttp.Request) {
	p := principalFrom(r.Context())
	h.writeJSON(w, nethttp.StatusOK, map[string]string{"role": string(p.Role), "subjectId": p.SubjectID})
}

// ---- испытуемые ----

func (h *Handler) publicSubjects(w nethttp.ResponseWriter, r *nethttp.Request) {
	list, err := h.subjects.ListPublic(r.Context())
	if err != nil {
		h.writeError(w, err)
		return
	}
	out := make([]publicSubjectDTO, 0, len(list))
	for _, s := range list {
		out = append(out, publicSubjectDTO{ID: s.ID, FullName: s.FullName, Group: s.Group})
	}
	h.writeJSON(w, nethttp.StatusOK, out)
}

func (h *Handler) listSubjects(w nethttp.ResponseWriter, r *nethttp.Request) {
	list, err := h.subjects.List(r.Context(), principalFrom(r.Context()), r.URL.Query().Get("group"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	out := make([]subjectDTO, 0, len(list))
	for _, s := range list {
		out = append(out, toSubjectDTO(s))
	}
	h.writeJSON(w, nethttp.StatusOK, out)
}

func (h *Handler) listGroups(w nethttp.ResponseWriter, r *nethttp.Request) {
	groups, err := h.subjects.Groups(r.Context(), principalFrom(r.Context()))
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeJSON(w, nethttp.StatusOK, groups)
}

func (h *Handler) getSubject(w nethttp.ResponseWriter, r *nethttp.Request) {
	s, err := h.subjects.Get(r.Context(), principalFrom(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeJSON(w, nethttp.StatusOK, toSubjectDTO(*s))
}

func (h *Handler) createSubject(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req subjectRequest
	if err := decode(r, &req); err != nil {
		h.writeError(w, err)
		return
	}
	s, err := h.subjects.Create(r.Context(), principalFrom(r.Context()), usecase.SubjectInput(req))
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeJSON(w, nethttp.StatusCreated, toSubjectDTO(*s))
}

func (h *Handler) updateSubject(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req subjectRequest
	if err := decode(r, &req); err != nil {
		h.writeError(w, err)
		return
	}
	s, err := h.subjects.Update(r.Context(), principalFrom(r.Context()), r.PathValue("id"), usecase.SubjectInput(req))
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeJSON(w, nethttp.StatusOK, toSubjectDTO(*s))
}

func (h *Handler) deleteSubject(w nethttp.ResponseWriter, r *nethttp.Request) {
	if err := h.subjects.Delete(r.Context(), principalFrom(r.Context()), r.PathValue("id")); err != nil {
		h.writeError(w, err)
		return
	}
	w.WriteHeader(nethttp.StatusNoContent)
}

// ---- настройки ----

func (h *Handler) getSettings(w nethttp.ResponseWriter, r *nethttp.Request) {
	s, err := h.settings.Get(r.Context())
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeJSON(w, nethttp.StatusOK, toSettingsDTO(s))
}

func (h *Handler) updateSettings(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req settingsDTO
	if err := decode(r, &req); err != nil {
		h.writeError(w, err)
		return
	}
	s, err := h.settings.Update(r.Context(), principalFrom(r.Context()), req.toDomain())
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeJSON(w, nethttp.StatusOK, toSettingsDTO(s))
}

// ---- тестирование и результаты ----

func (h *Handler) startTest(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req startTestRequest
	if err := decode(r, &req); err != nil {
		h.writeError(w, err)
		return
	}
	p := principalFrom(r.Context())
	if req.SubjectID == "" && p.Role == usecase.RoleSubject {
		req.SubjectID = p.SubjectID
	}
	var override *domain.Settings
	if req.Settings != nil {
		s := req.Settings.toDomain()
		override = &s
	}
	t, err := h.tests.Start(r.Context(), p, req.SubjectID, override)
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeJSON(w, nethttp.StatusCreated, toTestDTO(t))
}

func (h *Handler) finishTest(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req finishTestRequest
	if err := decode(r, &req); err != nil {
		h.writeError(w, err)
		return
	}
	results := make([]domain.TableResult, 0, len(req.Results))
	for _, x := range req.Results {
		results = append(results, domain.TableResult{Number: x.Number, Seconds: x.Seconds, Errors: x.Errors})
	}
	t, err := h.tests.Finish(r.Context(), principalFrom(r.Context()), r.PathValue("id"), results)
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeJSON(w, nethttp.StatusOK, toTestDTO(t))
}

func (h *Handler) listResults(w nethttp.ResponseWriter, r *nethttp.Request) {
	list, err := h.results.List(r.Context(), principalFrom(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	out := make([]testDTO, 0, len(list))
	for _, t := range list {
		out = append(out, toTestDTO(t))
	}
	h.writeJSON(w, nethttp.StatusOK, out)
}

func (h *Handler) getResult(w nethttp.ResponseWriter, r *nethttp.Request) {
	t, err := h.results.Get(r.Context(), principalFrom(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeJSON(w, nethttp.StatusOK, toTestDTO(t))
}

func (h *Handler) deleteResult(w nethttp.ResponseWriter, r *nethttp.Request) {
	if err := h.results.Delete(r.Context(), principalFrom(r.Context()), r.PathValue("id")); err != nil {
		h.writeError(w, err)
		return
	}
	w.WriteHeader(nethttp.StatusNoContent)
}

func (h *Handler) dynamics(w nethttp.ResponseWriter, r *nethttp.Request) {
	d, err := h.results.Dynamics(r.Context(), principalFrom(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeJSON(w, nethttp.StatusOK, toDynamicsDTO(d))
}

func (h *Handler) compare(w nethttp.ResponseWriter, r *nethttp.Request) {
	c, err := h.results.Compare(r.Context(), principalFrom(r.Context()), r.URL.Query().Get("a"), r.URL.Query().Get("b"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeJSON(w, nethttp.StatusOK, toComparisonDTO(c))
}

func (h *Handler) groupStats(w nethttp.ResponseWriter, r *nethttp.Request) {
	g, err := h.results.GroupStats(r.Context(), principalFrom(r.Context()), r.PathValue("name"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeJSON(w, nethttp.StatusOK, toGroupStatsDTO(g))
}
