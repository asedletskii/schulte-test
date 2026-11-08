package domain

import (
	"regexp"
)

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// размера шрифта в ячейках таблицы (в пикселях)
const (
	MinFontSize = 14
	MaxFontSize = 64
)

// визуальные параметры тестовых заданий
type Settings struct {
	FontSize        int
	BackgroundColor string
	ShowTimer       bool
	SoundOnError    bool
}

// возвращает стандартные параметры тестирования
func DefaultSettings() Settings {
	return Settings{
		FontSize:        28,
		BackgroundColor: "#ffffff",
		ShowTimer:       true,
		SoundOnError:    true,
	}
}

// проверяет корректность параметров
func (s Settings) Validate() error {
	if s.FontSize < MinFontSize || s.FontSize > MaxFontSize {
		return Validationf("размер шрифта должен быть в диапазоне %d–%d px", MinFontSize, MaxFontSize)
	}
	if !hexColor.MatchString(s.BackgroundColor) {
		return Validationf("цвет фона должен быть задан в формате #RRGGBB")
	}
	return nil
}
