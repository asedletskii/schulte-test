package domain

import (
	"math"
	"sort"
	"time"
)

// DynamicsPoint — точка динамики результатов испытуемого.
type DynamicsPoint struct {
	SessionID      string
	Date           time.Time
	WorkEfficiency float64
	WarmUp         float64
	Stability      float64
	Errors         int
}

// Trend — направление изменения показателя.
type Trend string

const (
	TrendImproving Trend = "improving" // время уменьшается
	TrendWorsening Trend = "worsening" // время растёт
	TrendStable    Trend = "stable"
	TrendUnknown   Trend = "unknown" // недостаточно данных (менее двух тестов)
)

// trendThreshold — минимальное изменение ЭР за один тест (в секундах),
// начиная с которого тренд считается значимым.
const trendThreshold = 0.05

// Dynamics — отчёт об изменении характеристик внимания во времени.
type Dynamics struct {
	Points []DynamicsPoint
	Slope  float64 // наклон линейной регрессии ЭР по номеру теста, с/тест
	Trend  Trend
}

// BuildDynamics строит динамику по завершённым сеансам (в хронологическом порядке).
func BuildDynamics(sessions []*TestSession) Dynamics {
	done := completed(sessions)
	sort.Slice(done, func(i, j int) bool { return done[i].FinishedAt.Before(*done[j].FinishedAt) })

	d := Dynamics{Points: make([]DynamicsPoint, 0, len(done)), Trend: TrendUnknown}
	for _, s := range done {
		d.Points = append(d.Points, DynamicsPoint{
			SessionID:      s.ID,
			Date:           *s.FinishedAt,
			WorkEfficiency: s.Metrics.WorkEfficiency,
			WarmUp:         s.Metrics.WarmUp,
			Stability:      s.Metrics.Stability,
			Errors:         s.Metrics.TotalErrors,
		})
	}
	if len(d.Points) < 2 {
		return d
	}
	ys := make([]float64, len(d.Points))
	for i, p := range d.Points {
		ys[i] = p.WorkEfficiency
	}
	d.Slope = linearSlope(ys)
	switch {
	case d.Slope < -trendThreshold:
		d.Trend = TrendImproving
	case d.Slope > trendThreshold:
		d.Trend = TrendWorsening
	default:
		d.Trend = TrendStable
	}
	return d
}

// linearSlope — наклон МНК-прямой y = a + b·x при x = 0, 1, 2, …
func linearSlope(ys []float64) float64 {
	n := float64(len(ys))
	var sx, sy, sxy, sxx float64
	for i, y := range ys {
		x := float64(i)
		sx += x
		sy += y
		sxy += x * y
		sxx += x * x
	}
	den := n*sxx - sx*sx
	if den == 0 {
		return 0
	}
	return (n*sxy - sx*sy) / den
}

// Comparison — сравнение двух сеансов тестирования.
type Comparison struct {
	A, B           *TestSession
	DiffEfficiency float64 // ЭР(B) − ЭР(A), с; отрицательное — B быстрее
	DiffWarmUp     float64
	DiffStability  float64
	DiffErrors     int
	// Better — «A», «B» или «equal»: у кого ниже ЭР (меньше времени на таблицу).
	Better string
}

// Compare сравнивает два завершённых сеанса.
func Compare(a, b *TestSession) (Comparison, error) {
	if a == nil || b == nil || a.Metrics == nil || b.Metrics == nil {
		return Comparison{}, Validationf("для сравнения нужны два завершённых тестирования")
	}
	c := Comparison{
		A: a, B: b,
		DiffEfficiency: b.Metrics.WorkEfficiency - a.Metrics.WorkEfficiency,
		DiffWarmUp:     b.Metrics.WarmUp - a.Metrics.WarmUp,
		DiffStability:  b.Metrics.Stability - a.Metrics.Stability,
		DiffErrors:     b.Metrics.TotalErrors - a.Metrics.TotalErrors,
	}
	switch {
	case math.Abs(c.DiffEfficiency) < 1e-9:
		c.Better = "equal"
	case c.DiffEfficiency < 0:
		c.Better = "B"
	default:
		c.Better = "A"
	}
	return c, nil
}

// GroupMember — итог испытуемого в групповом сравнении (по последнему тесту).
type GroupMember struct {
	Subject Subject
	Tests   int
	Latest  *Metrics // nil, если нет завершённых тестов
	Rank    int      // место по ЭР (1 — лучший); 0 — нет данных
}

// GroupStats — сводка по группе испытуемых.
type GroupStats struct {
	Group           string
	Members         []GroupMember
	AvgEfficiency   float64
	AvgWarmUp       float64
	AvgStability    float64
	ParticipantsCnt int // число участников, имеющих результаты
}

// BuildGroupStats формирует сводку: средние значения по последним тестам и рейтинг.
func BuildGroupStats(group string, subjects []Subject, sessionsBySubject map[string][]*TestSession) GroupStats {
	gs := GroupStats{Group: group}
	for _, sub := range subjects {
		done := completed(sessionsBySubject[sub.ID])
		sort.Slice(done, func(i, j int) bool { return done[i].FinishedAt.Before(*done[j].FinishedAt) })
		m := GroupMember{Subject: sub, Tests: len(done)}
		if len(done) > 0 {
			m.Latest = done[len(done)-1].Metrics
		}
		gs.Members = append(gs.Members, m)
	}
	// рейтинг по возрастанию ЭР
	idx := make([]int, 0, len(gs.Members))
	for i, m := range gs.Members {
		if m.Latest != nil {
			idx = append(idx, i)
		}
	}
	sort.SliceStable(idx, func(a, b int) bool {
		return gs.Members[idx[a]].Latest.WorkEfficiency < gs.Members[idx[b]].Latest.WorkEfficiency
	})
	for rank, i := range idx {
		gs.Members[i].Rank = rank + 1
		gs.AvgEfficiency += gs.Members[i].Latest.WorkEfficiency
		gs.AvgWarmUp += gs.Members[i].Latest.WarmUp
		gs.AvgStability += gs.Members[i].Latest.Stability
	}
	gs.ParticipantsCnt = len(idx)
	if n := float64(len(idx)); n > 0 {
		gs.AvgEfficiency /= n
		gs.AvgWarmUp /= n
		gs.AvgStability /= n
	}
	return gs
}

func completed(sessions []*TestSession) []*TestSession {
	out := make([]*TestSession, 0, len(sessions))
	for _, s := range sessions {
		if s != nil && s.Status == StatusCompleted && s.Metrics != nil && s.FinishedAt != nil {
			out = append(out, s)
		}
	}
	return out
}
