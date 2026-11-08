package domain

import (
	"strings"
	"time"
	"unicode/utf8"
)

// испытуемый, проходящий тестирование
type Subject struct {
	ID        string
	FullName  string
	BirthYear int    // 0 — не указан
	Group     string // учебная/рабочая группа, используется для сравнения в группе
	Notes     string
	CreatedAt time.Time
}

// проверяет инварианты сущности
func (s *Subject) Validate(now time.Time) error {
	s.FullName = strings.TrimSpace(s.FullName)
	s.Group = strings.TrimSpace(s.Group)
	s.Notes = strings.TrimSpace(s.Notes)

	if n := utf8.RuneCountInString(s.FullName); n < 2 || n > 120 {
		return Validationf("ФИО должно содержать от 2 до 120 символов")
	}
	if utf8.RuneCountInString(s.Group) > 60 {
		return Validationf("название группы не длиннее 60 символов")
	}
	if utf8.RuneCountInString(s.Notes) > 500 {
		return Validationf("примечание не длиннее 500 символов")
	}
	if s.BirthYear != 0 && (s.BirthYear < 1900 || s.BirthYear > now.Year()) {
		return Validationf("год рождения должен быть в диапазоне 1900–%d", now.Year())
	}
	return nil
}
