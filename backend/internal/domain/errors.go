package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound     = errors.New("объект не найден")
	ErrValidation   = errors.New("некорректные данные")
	ErrUnauthorized = errors.New("требуется авторизация")
	ErrForbidden    = errors.New("недостаточно прав")
	ErrConflict     = errors.New("операция невозможна в текущем состоянии")
)

func Validationf(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrValidation, fmt.Sprintf(format, args...))
}
