package domain

import (
	"math/rand/v2"
	"sync"
)

const (
	// GridSize — сторона таблицы Шульте.
	GridSize = 5
	// CellsCount — количество чисел в одной таблице (1…25).
	CellsCount = GridSize * GridSize
	// TablesCount — количество таблиц в одном тестировании по методике.
	TablesCount = 5
)

// TableGenerator формирует таблицы Шульте со случайным расположением чисел.
type TableGenerator struct {
	mu  sync.Mutex // rand.Rand не потокобезопасен, а сервер обслуживает запросы параллельно
	rnd *rand.Rand
}

// NewTableGenerator создаёт генератор. Передав фиксированный seed, можно
// получить воспроизводимые таблицы (используется в тестах).
func NewTableGenerator(seed1, seed2 uint64) *TableGenerator {
	return &TableGenerator{rnd: rand.New(rand.NewPCG(seed1, seed2))}
}

// Generate возвращает случайную перестановку чисел от 1 до 25.
// Индекс элемента — позиция ячейки слева направо, сверху вниз.
func (g *TableGenerator) Generate() []int {
	cells := make([]int, CellsCount)
	for i := range cells {
		cells[i] = i + 1
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.rnd.Shuffle(len(cells), func(i, j int) { cells[i], cells[j] = cells[j], cells[i] })
	return cells
}

// GenerateAll возвращает полный комплект из пяти таблиц.
func (g *TableGenerator) GenerateAll() [][]int {
	tables := make([][]int, TablesCount)
	for i := range tables {
		tables[i] = g.Generate()
	}
	return tables
}

// IsValidTable проверяет, что таблица — перестановка чисел 1…25.
func IsValidTable(cells []int) bool {
	if len(cells) != CellsCount {
		return false
	}
	seen := make([]bool, CellsCount+1)
	for _, v := range cells {
		if v < 1 || v > CellsCount || seen[v] {
			return false
		}
		seen[v] = true
	}
	return true
}
