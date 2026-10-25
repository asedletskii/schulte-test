package domain

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"
)

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func results(secs ...float64) []TableResult {
	out := make([]TableResult, len(secs))
	for i, s := range secs {
		out[i] = TableResult{Number: i + 1, Seconds: s}
	}
	return out
}

func TestCalculateMetrics(t *testing.T) {
	// T = 40, 38, 36, 44, 42 → ЭР = 200/5 = 40; ВР = 40/40 = 1,0; ПУ = 44/40 = 1,1
	m, err := CalculateMetrics(results(40, 38, 36, 44, 42))
	if err != nil {
		t.Fatal(err)
	}
	if !approx(m.WorkEfficiency, 40) || !approx(m.WarmUp, 1.0) || !approx(m.Stability, 1.1) {
		t.Fatalf("неверные показатели: %+v", m)
	}
	if m.WarmUpLevel != LevelGood {
		t.Errorf("ВР = 1,0 должна считаться хорошей, получено %s", m.WarmUpLevel)
	}
	if m.StabilityLevel != LevelBad {
		t.Errorf("ПУ = 1,1 должна считаться низкой, получено %s", m.StabilityLevel)
	}
	if !approx(m.TotalSeconds, 200) || len(m.FatigueCurve) != 5 {
		t.Errorf("суммарное время/кривая утомляемости: %+v", m)
	}
}

func TestCalculateMetricsGoodWarmUp(t *testing.T) {
	m, _ := CalculateMetrics(results(30, 35, 40, 30, 45)) // ЭР=36; ВР≈0,833; ПУ≈0,833
	if m.WarmUpLevel != LevelGood || m.StabilityLevel != LevelGood {
		t.Fatalf("ожидалась хорошая оценка: %+v", m)
	}
}

func TestCalculateMetricsErrors(t *testing.T) {
	if _, err := CalculateMetrics(results(1, 2, 3)); !errors.Is(err, ErrValidation) {
		t.Error("ожидалась ошибка при менее чем пяти таблицах")
	}
	if _, err := CalculateMetrics(results(1, 2, 0, 4, 5)); !errors.Is(err, ErrValidation) {
		t.Error("ожидалась ошибка при нулевом времени")
	}
}

func TestTableGenerator(t *testing.T) {
	g := NewTableGenerator(1, 2)
	tables := g.GenerateAll()
	if len(tables) != TablesCount {
		t.Fatalf("ожидалось %d таблиц", TablesCount)
	}
	for i, tb := range tables {
		if !IsValidTable(tb) {
			t.Errorf("таблица %d не является перестановкой 1…25: %v", i, tb)
		}
	}
	// воспроизводимость при одинаковом seed
	again := NewTableGenerator(1, 2).GenerateAll()
	for i := range tables {
		for j := range tables[i] {
			if tables[i][j] != again[i][j] {
				t.Fatal("одинаковый seed должен давать одинаковые таблицы")
			}
		}
	}
	// таблицы в комплекте различаются
	same := true
	for j := range tables[0] {
		if tables[0][j] != tables[1][j] {
			same = false
		}
	}
	if same {
		t.Error("две подряд сгенерированные таблицы совпали")
	}
}

func TestIsValidTable(t *testing.T) {
	bad := make([]int, CellsCount)
	for i := range bad {
		bad[i] = 1
	}
	if IsValidTable(bad) || IsValidTable([]int{1, 2, 3}) {
		t.Error("некорректные таблицы приняты")
	}
}

func TestSubjectValidate(t *testing.T) {
	now := time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC)
	ok := Subject{FullName: "  Иванов Иван  ", BirthYear: 2000, Group: "ПМИ-61"}
	if err := ok.Validate(now); err != nil || ok.FullName != "Иванов Иван" {
		t.Fatalf("корректный испытуемый отклонён: %v / %q", err, ok.FullName)
	}
	cases := map[string]Subject{
		"пустое ФИО":      {FullName: " "},
		"длинное ФИО":     {FullName: strings.Repeat("а", 121)},
		"год из будущего": {FullName: "Иван", BirthYear: 2999},
		"слишком давно":   {FullName: "Иван", BirthYear: 1800},
	}
	for name, s := range cases {
		if err := s.Validate(now); !errors.Is(err, ErrValidation) {
			t.Errorf("%s: ожидалась ошибка валидации, получено %v", name, err)
		}
	}
}

func TestSettingsValidate(t *testing.T) {
	if err := DefaultSettings().Validate(); err != nil {
		t.Fatal(err)
	}
	for _, s := range []Settings{
		{FontSize: 5, BackgroundColor: "#ffffff"},
		{FontSize: 100, BackgroundColor: "#ffffff"},
		{FontSize: 20, BackgroundColor: "white"},
	} {
		if err := s.Validate(); !errors.Is(err, ErrValidation) {
			t.Errorf("настройки %+v должны быть отклонены", s)
		}
	}
}

func newSession(t *testing.T) *TestSession {
	t.Helper()
	s, err := NewTestSession("s1", "u1", NewTableGenerator(3, 4).GenerateAll(), DefaultSettings(), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestSessionComplete(t *testing.T) {
	s := newSession(t)
	now := time.Now()
	if err := s.Complete(results(40, 38, 36, 44, 42), now); err != nil {
		t.Fatal(err)
	}
	if s.Status != StatusCompleted || s.Metrics == nil || s.FinishedAt == nil {
		t.Fatalf("сеанс не завершён: %+v", s)
	}
	if err := s.Complete(results(1, 2, 3, 4, 5), now); !errors.Is(err, ErrConflict) {
		t.Error("повторное завершение должно давать ErrConflict")
	}
}

func TestSessionCompleteValidation(t *testing.T) {
	now := time.Now()
	for name, r := range map[string][]TableResult{
		"мало таблиц":   results(1, 2),
		"нулевое время": results(1, 2, 0, 4, 5),
		"слишком долго": results(1, 2, 3, 4, MaxTableSeconds+1),
		"отрицательные": {{1, 10, -1}, {2, 10, 0}, {3, 10, 0}, {4, 10, 0}, {5, 10, 0}},
	} {
		if err := newSession(t).Complete(r, now); !errors.Is(err, ErrValidation) {
			t.Errorf("%s: ожидалась ошибка валидации, получено %v", name, err)
		}
	}
}

func completedSession(t *testing.T, id string, at time.Time, secs ...float64) *TestSession {
	t.Helper()
	s := newSession(t)
	s.ID = id
	if err := s.Complete(results(secs...), at); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestDynamics(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	sessions := []*TestSession{
		completedSession(t, "c", base.Add(48*time.Hour), 30, 30, 30, 30, 30), // ЭР=30
		completedSession(t, "a", base, 50, 50, 50, 50, 50),                   // ЭР=50
		completedSession(t, "b", base.Add(24*time.Hour), 40, 40, 40, 40, 40), // ЭР=40
	}
	d := BuildDynamics(sessions)
	if len(d.Points) != 3 || d.Points[0].SessionID != "a" || d.Points[2].SessionID != "c" {
		t.Fatalf("точки должны идти в хронологическом порядке: %+v", d.Points)
	}
	if !approx(d.Slope, -10) || d.Trend != TrendImproving {
		t.Errorf("ожидалось улучшение с наклоном −10, получено %v / %s", d.Slope, d.Trend)
	}
	if BuildDynamics(sessions[:1]).Trend != TrendUnknown {
		t.Error("для одного теста тренд неизвестен")
	}
	flat := BuildDynamics([]*TestSession{
		completedSession(t, "x", base, 40, 40, 40, 40, 40),
		completedSession(t, "y", base.Add(time.Hour), 40, 40, 40, 40, 40),
	})
	if flat.Trend != TrendStable {
		t.Errorf("ожидался стабильный тренд, получено %s", flat.Trend)
	}
}

func TestCompare(t *testing.T) {
	now := time.Now()
	a := completedSession(t, "a", now, 50, 50, 50, 50, 50)
	b := completedSession(t, "b", now, 40, 40, 40, 40, 40)
	c, err := Compare(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if c.Better != "B" || !approx(c.DiffEfficiency, -10) {
		t.Errorf("быстрее должен быть B: %+v", c)
	}
	if c2, _ := Compare(a, a); c2.Better != "equal" {
		t.Error("одинаковые тесты должны давать equal")
	}
	if _, err := Compare(a, nil); !errors.Is(err, ErrValidation) {
		t.Error("ожидалась ошибка для nil")
	}
}

func TestGroupStats(t *testing.T) {
	now := time.Now()
	subs := []Subject{{ID: "1", FullName: "А"}, {ID: "2", FullName: "Б"}, {ID: "3", FullName: "В"}}
	by := map[string][]*TestSession{
		"1": {completedSession(t, "s1", now, 50, 50, 50, 50, 50)},
		"2": {completedSession(t, "s2", now, 30, 30, 30, 30, 30)},
	}
	g := BuildGroupStats("G", subs, by)
	if g.ParticipantsCnt != 2 || !approx(g.AvgEfficiency, 40) {
		t.Fatalf("сводка по группе: %+v", g)
	}
	if g.Members[1].Rank != 1 || g.Members[0].Rank != 2 || g.Members[2].Rank != 0 {
		t.Errorf("рейтинг: %+v", g.Members)
	}
}
