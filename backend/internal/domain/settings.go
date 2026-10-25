package domain

import (
	"regexp"
)

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// Границы допустимого размера шрифта в ячейках таблицы (в пикселях).
const (
	MinFontSize = 14
	MaxFontSize = 64
)

// Settings — визуальные параметры тестовых заданий (требование ТЗ 1.3).
type Settings struct {
	FontSize        int    // размер шрифта чисел в таблице, px
	BackgroundColor string // цвет фона таблицы, #RRGGBB
	ShowTimer       bool   // показывать секундомер во время теста
	SoundOnError    bool   // звуковой сигнал при ошибочном выборе
}

// DefaultSettings возвращает стандартные параметры тестирования.
func DefaultSettings() Settings {
	return Settings{
		FontSize:        28,
		BackgroundColor: "#ffffff",
		ShowTimer:       true,
		SoundOnError:    true,
	}
}

// Validate проверяет корректность параметров.
func (s Settings) Validate() error {
	if s.FontSize < MinFontSize || s.FontSize > MaxFontSize {
		return Validationf("размер шрифта должен быть в диапазоне %d–%d px", MinFontSize, MaxFontSize)
	}
	if !hexColor.MatchString(s.BackgroundColor) {
		return Validationf("цвет фона должен быть задан в формате #RRGGBB")
	}
	return nil
}
