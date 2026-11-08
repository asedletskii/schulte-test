package http

import (
	"log/slog"
	nethttp "net/http"
)

// собирает маршруты REST API и подключает сквозные обработчики
func NewRouter(h *Handler, log *slog.Logger, corsOrigin string) nethttp.Handler {
	mux := nethttp.NewServeMux()
	a := h.authenticated

	mux.HandleFunc("GET /api/health", func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		h.writeJSON(w, nethttp.StatusOK, map[string]string{"status": "ok"})
	})

	// Вход в систему (без токена)
	mux.HandleFunc("POST /api/auth/admin", h.loginAdmin)
	mux.HandleFunc("POST /api/auth/subject", h.loginSubject)
	mux.HandleFunc("GET /api/public/subjects", h.publicSubjects)
	mux.HandleFunc("GET /api/auth/me", a(h.me))

	// Испытуемые
	mux.HandleFunc("GET /api/subjects", a(h.listSubjects))
	mux.HandleFunc("POST /api/subjects", a(h.createSubject))
	mux.HandleFunc("GET /api/subjects/{id}", a(h.getSubject))
	mux.HandleFunc("PUT /api/subjects/{id}", a(h.updateSubject))
	mux.HandleFunc("DELETE /api/subjects/{id}", a(h.deleteSubject))
	mux.HandleFunc("GET /api/subjects/{id}/results", a(h.listResults))
	mux.HandleFunc("GET /api/subjects/{id}/dynamics", a(h.dynamics))

	// Группы
	mux.HandleFunc("GET /api/groups", a(h.listGroups))
	mux.HandleFunc("GET /api/groups/{name}/stats", a(h.groupStats))

	// Настройки
	mux.HandleFunc("GET /api/settings", a(h.getSettings))
	mux.HandleFunc("PUT /api/settings", a(h.updateSettings))

	// Тестирование и результаты
	mux.HandleFunc("POST /api/tests", a(h.startTest))
	mux.HandleFunc("POST /api/tests/{id}/finish", a(h.finishTest))
	mux.HandleFunc("GET /api/results/{id}", a(h.getResult))
	mux.HandleFunc("DELETE /api/results/{id}", a(h.deleteResult))
	mux.HandleFunc("GET /api/compare", a(h.compare))

	return logging(log, cors(corsOrigin, mux))
}
