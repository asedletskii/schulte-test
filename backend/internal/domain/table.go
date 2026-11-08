package domain

import (
	"math/rand/v2"
	"sync"
)

const (
	GridSize = 5 // сторона таблицы Шульте
	CellsCount = GridSize * GridSize // количество чисел в одной таблице
	TablesCount = 5	// количество таблиц в одном тестировании по методике
)

// формирует таблицы Шульте со случайным расположением чисел
type TableGenerator struct {
	mu  sync.Mutex
	rnd *rand.Rand
}

// создаёт генератор. Передав фиксированный seed, чтобы получить воспроизводимые таблицы (в тестах)
func NewTableGenerator(seed1, seed2 uint64) *TableGenerator {
	return &TableGenerator{rnd: rand.New(rand.NewPCG(seed1, seed2))}
}

// возвращает случайную перестановку чисел от 1 до 25
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

// возвращает полный комплект из пяти таблиц
func (g *TableGenerator) GenerateAll() [][]int {
	tables := make([][]int, TablesCount)
	for i := range tables {
		tables[i] = g.Generate()
	}
	return tables
}

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
