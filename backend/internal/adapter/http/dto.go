// Пакет http — адаптер доставки (REST API на net/http). Преобразует
// HTTP-запросы в вызовы прикладных сервисов и обратно.
package http

import (
	"time"

	"schulte/internal/domain"
)

// ---- Запросы ----

type adminLoginRequest struct {
	Password string `json:"password"`
}

type subjectLoginRequest struct {
	SubjectID string `json:"subjectId"`
}

type subjectRequest struct {
	FullName  string `json:"fullName"`
	BirthYear int    `json:"birthYear"`
	Group     string `json:"group"`
	Notes     string `json:"notes"`
}

type startTestRequest struct {
	SubjectID string       `json:"subjectId"`
	Settings  *settingsDTO `json:"settings,omitempty"`
}

type finishTestRequest struct {
	Results []tableResultDTO `json:"results"`
}

// ---- Ответы ----

type sessionResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	Role      string    `json:"role"`
	SubjectID string    `json:"subjectId,omitempty"`
}

type subjectDTO struct {
	ID        string    `json:"id"`
	FullName  string    `json:"fullName"`
	BirthYear int       `json:"birthYear"`
	Group     string    `json:"group"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"createdAt"`
}

type publicSubjectDTO struct {
	ID       string `json:"id"`
	FullName string `json:"fullName"`
	Group    string `json:"group"`
}

type settingsDTO struct {
	FontSize        int    `json:"fontSize"`
	BackgroundColor string `json:"backgroundColor"`
	ShowTimer       bool   `json:"showTimer"`
	SoundOnError    bool   `json:"soundOnError"`
}

type tableResultDTO struct {
	Number  int     `json:"number"`
	Seconds float64 `json:"seconds"`
	Errors  int     `json:"errors"`
}

type metricsDTO struct {
	TotalSeconds   float64   `json:"totalSeconds"`
	WorkEfficiency float64   `json:"workEfficiency"`
	WarmUp         float64   `json:"warmUp"`
	Stability      float64   `json:"stability"`
	WarmUpLevel    string    `json:"warmUpLevel"`
	StabilityLevel string    `json:"stabilityLevel"`
	FatigueCurve   []float64 `json:"fatigueCurve"`
	TotalErrors    int       `json:"totalErrors"`
}

type testDTO struct {
	ID         string           `json:"id"`
	SubjectID  string           `json:"subjectId"`
	Status     string           `json:"status"`
	Tables     [][]int          `json:"tables,omitempty"` // только для незавершённого теста
	Settings   settingsDTO      `json:"settings"`
	StartedAt  time.Time        `json:"startedAt"`
	FinishedAt *time.Time       `json:"finishedAt,omitempty"`
	Results    []tableResultDTO `json:"results,omitempty"`
	Metrics    *metricsDTO      `json:"metrics,omitempty"`
}

type dynamicsPointDTO struct {
	SessionID      string    `json:"sessionId"`
	Date           time.Time `json:"date"`
	WorkEfficiency float64   `json:"workEfficiency"`
	WarmUp         float64   `json:"warmUp"`
	Stability      float64   `json:"stability"`
	Errors         int       `json:"errors"`
}

type dynamicsDTO struct {
	Points []dynamicsPointDTO `json:"points"`
	Slope  float64            `json:"slope"`
	Trend  string             `json:"trend"`
}

type comparisonDTO struct {
	A              testDTO `json:"a"`
	B              testDTO `json:"b"`
	DiffEfficiency float64 `json:"diffEfficiency"`
	DiffWarmUp     float64 `json:"diffWarmUp"`
	DiffStability  float64 `json:"diffStability"`
	DiffErrors     int     `json:"diffErrors"`
	Better         string  `json:"better"`
}

type groupMemberDTO struct {
	Subject subjectDTO  `json:"subject"`
	Tests   int         `json:"tests"`
	Latest  *metricsDTO `json:"latest,omitempty"`
	Rank    int         `json:"rank"`
}

type groupStatsDTO struct {
	Group         string           `json:"group"`
	Members       []groupMemberDTO `json:"members"`
	AvgEfficiency float64          `json:"avgEfficiency"`
	AvgWarmUp     float64          `json:"avgWarmUp"`
	AvgStability  float64          `json:"avgStability"`
	Participants  int              `json:"participants"`
}

// ---- Преобразования domain → DTO ----

func toSubjectDTO(s domain.Subject) subjectDTO {
	return subjectDTO{ID: s.ID, FullName: s.FullName, BirthYear: s.BirthYear, Group: s.Group, Notes: s.Notes, CreatedAt: s.CreatedAt}
}

func toSettingsDTO(s domain.Settings) settingsDTO {
	return settingsDTO{FontSize: s.FontSize, BackgroundColor: s.BackgroundColor, ShowTimer: s.ShowTimer, SoundOnError: s.SoundOnError}
}

func (s settingsDTO) toDomain() domain.Settings {
	return domain.Settings{FontSize: s.FontSize, BackgroundColor: s.BackgroundColor, ShowTimer: s.ShowTimer, SoundOnError: s.SoundOnError}
}

func toMetricsDTO(m *domain.Metrics) *metricsDTO {
	if m == nil {
		return nil
	}
	return &metricsDTO{
		TotalSeconds: m.TotalSeconds, WorkEfficiency: m.WorkEfficiency, WarmUp: m.WarmUp, Stability: m.Stability,
		WarmUpLevel: m.WarmUpLevel, StabilityLevel: m.StabilityLevel, FatigueCurve: m.FatigueCurve, TotalErrors: m.TotalErrors,
	}
}

func toTestDTO(t *domain.TestSession) testDTO {
	d := testDTO{
		ID: t.ID, SubjectID: t.SubjectID, Status: string(t.Status), Settings: toSettingsDTO(t.Settings),
		StartedAt: t.StartedAt, FinishedAt: t.FinishedAt, Metrics: toMetricsDTO(t.Metrics),
	}
	if t.Status == domain.StatusInProgress {
		d.Tables = t.Tables
	}
	for _, r := range t.Results {
		d.Results = append(d.Results, tableResultDTO{Number: r.Number, Seconds: r.Seconds, Errors: r.Errors})
	}
	return d
}

func toDynamicsDTO(d domain.Dynamics) dynamicsDTO {
	out := dynamicsDTO{Points: make([]dynamicsPointDTO, 0, len(d.Points)), Slope: d.Slope, Trend: string(d.Trend)}
	for _, p := range d.Points {
		out.Points = append(out.Points, dynamicsPointDTO{
			SessionID: p.SessionID, Date: p.Date, WorkEfficiency: p.WorkEfficiency,
			WarmUp: p.WarmUp, Stability: p.Stability, Errors: p.Errors,
		})
	}
	return out
}

func toComparisonDTO(c domain.Comparison) comparisonDTO {
	return comparisonDTO{
		A: toTestDTO(c.A), B: toTestDTO(c.B), DiffEfficiency: c.DiffEfficiency, DiffWarmUp: c.DiffWarmUp,
		DiffStability: c.DiffStability, DiffErrors: c.DiffErrors, Better: c.Better,
	}
}

func toGroupStatsDTO(g domain.GroupStats) groupStatsDTO {
	out := groupStatsDTO{
		Group: g.Group, Members: make([]groupMemberDTO, 0, len(g.Members)), AvgEfficiency: g.AvgEfficiency,
		AvgWarmUp: g.AvgWarmUp, AvgStability: g.AvgStability, Participants: g.ParticipantsCnt,
	}
	for _, m := range g.Members {
		out.Members = append(out.Members, groupMemberDTO{
			Subject: toSubjectDTO(m.Subject), Tests: m.Tests, Latest: toMetricsDTO(m.Latest), Rank: m.Rank,
		})
	}
	return out
}
