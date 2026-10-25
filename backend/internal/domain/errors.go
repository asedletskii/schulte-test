// Пакет domain содержит сущности предметной области «Тестирование внимания
// по методике Шульте» и бизнес-правила, не зависящие от фреймворков, БД и
// транспорта. Это внутренний (центральный) слой чистой архитектуры.
package domain

import (
	"errors"
	"fmt"
)

// Базовые ошибки предметной области. Внешние слои сопоставляют их с
// HTTP-статусами, не зная деталей реализации.
var (
	ErrNotFound     = errors.New("объект не найден")
	ErrValidation   = errors.New("некорректные данные")
	ErrUnauthorized = errors.New("требуется авторизация")
	ErrForbidden    = errors.New("недостаточно прав")
	ErrConflict     = errors.New("операция невозможна в текущем состоянии")
)

// Validationf формирует ошибку валидации с пояснением.
func Validationf(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrValidation, fmt.Sprintf(format, args...))
}
