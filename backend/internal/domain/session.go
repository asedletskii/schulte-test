package domain

import "time"

// SessionStatus — состояние сеанса тестирования.
type SessionStatus string

const (
	StatusInProgress SessionStatus = "in_progress"
	StatusCompleted  SessionStatus = "completed"
)

// MaxTableSeconds — верхняя граница времени на одну таблицу (защита от
// заведомо некорректных данных клиента).
const MaxTableSeconds = 3600.0

// TableResult — итог прохождения одной таблицы.
type TableResult struct {
	Number  int     // порядковый номер таблицы 1…5
	Seconds float64 // затраченное время, с
	Errors  int     // число ошибочных выборов
}

// TestSession — один сеанс тестирования испытуемого (5 таблиц).
type TestSession struct {
	ID         string
	SubjectID  string
	Status     SessionStatus
	Tables     [][]int  // сгенерированные таблицы
	Settings   Settings // параметры на момент проведения теста
	StartedAt  time.Time
	FinishedAt *time.Time
	Results    []TableResult
	Metrics    *Metrics
}

// NewTestSession создаёт сеанс в состоянии «выполняется».
func NewTestSession(id, subjectID string, tables [][]int, s Settings, now time.Time) (*TestSession, error) {
	if len(tables) != TablesCount {
		return nil, Validationf("ожидается %d таблиц", TablesCount)
	}
	for i, t := range tables {
		if !IsValidTable(t) {
			return nil, Validationf("таблица %d не является перестановкой чисел 1…%d", i+1, CellsCount)
		}
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}
	return &TestSession{
		ID:        id,
		SubjectID: subjectID,
		Status:    StatusInProgress,
		Tables:    tables,
		Settings:  s,
		StartedAt: now,
	}, nil
}

// Complete фиксирует результаты и рассчитывает показатели внимания.
func (t *TestSession) Complete(results []TableResult, now time.Time) error {
	if t.Status == StatusCompleted {
		return ErrConflict
	}
	if len(results) != TablesCount {
		return Validationf("необходимо передать результаты по %d таблицам", TablesCount)
	}
	normalized := make([]TableResult, len(results))
	for i, r := range results {
		if r.Seconds <= 0 || r.Seconds > MaxTableSeconds {
			return Validationf("время по таблице %d должно быть в диапазоне (0; %.0f] с", i+1, MaxTableSeconds)
		}
		if r.Errors < 0 || r.Errors > 1000 {
			return Validationf("число ошибок по таблице %d указано неверно", i+1)
		}
		normalized[i] = TableResult{Number: i + 1, Seconds: r.Seconds, Errors: r.Errors}
	}
	m, err := CalculateMetrics(normalized)
	if err != nil {
		return err
	}
	t.Results = normalized
	t.Metrics = &m
	t.Status = StatusCompleted
	t.FinishedAt = &now
	return nil
}
