package domain

// Интерпретация показателей (см. практикум, формулы 1.1–1.3).
const (
	LevelGood = "good" // показатель не выше 1,0
	LevelBad  = "bad"  // показатель выше 1,0
)

// Metrics — рассчитанные показатели внимания по методике Шульте.
type Metrics struct {
	TotalSeconds   float64 // суммарное время по всем таблицам, с
	WorkEfficiency float64 // ЭР — эффективность работы, с (формула 1.1)
	WarmUp         float64 // ВР — степень врабатываемости (формула 1.2)
	Stability      float64 // ПУ — психическая устойчивость (формула 1.3)
	WarmUpLevel    string  // LevelGood: ВР ≤ 1,0 — хорошая врабатываемость
	StabilityLevel string  // LevelGood: ПУ ≤ 1,0 — хорошая устойчивость
	FatigueCurve   []float64
	TotalErrors    int
}

// CalculateMetrics вычисляет показатели по времени прохождения пяти таблиц.
//
//	ЭР = (T1+T2+T3+T4+T5) / 5
//	ВР = T1 / ЭР
//	ПУ = T4 / ЭР
func CalculateMetrics(results []TableResult) (Metrics, error) {
	if len(results) != TablesCount {
		return Metrics{}, Validationf("для расчёта требуется время по %d таблицам", TablesCount)
	}
	var sum float64
	var errs int
	curve := make([]float64, TablesCount)
	for i, r := range results {
		if r.Seconds <= 0 {
			return Metrics{}, Validationf("время по таблице %d должно быть положительным", i+1)
		}
		sum += r.Seconds
		errs += r.Errors
		curve[i] = r.Seconds
	}
	er := sum / TablesCount
	m := Metrics{
		TotalSeconds:   sum,
		WorkEfficiency: er,
		WarmUp:         results[0].Seconds / er,
		Stability:      results[3].Seconds / er,
		FatigueCurve:   curve,
		TotalErrors:    errs,
	}
	m.WarmUpLevel = level(m.WarmUp)
	m.StabilityLevel = level(m.Stability)
	return m, nil
}

func level(v float64) string {
	if v <= 1.0 {
		return LevelGood
	}
	return LevelBad
}
