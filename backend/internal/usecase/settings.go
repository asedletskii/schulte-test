package usecase

import (
	"context"
	"errors"

	"schulte/internal/domain"
)

// SettingsService — вариант использования «Настроить параметры тестирования».
type SettingsService struct {
	repo SettingsRepository
}

// NewSettingsService создаёт сервис настроек.
func NewSettingsService(repo SettingsRepository) *SettingsService {
	return &SettingsService{repo: repo}
}

// Get возвращает текущие параметры (доступно любому аутентифицированному пользователю).
func (s *SettingsService) Get(ctx context.Context) (domain.Settings, error) {
	st, err := s.repo.Get(ctx)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.DefaultSettings(), nil // нет сохранённых — используем стандартные
	}
	if err != nil {
		return domain.Settings{}, err
	}
	return st, nil
}

// Update сохраняет параметры по умолчанию (только администратор).
func (s *SettingsService) Update(ctx context.Context, p Principal, st domain.Settings) (domain.Settings, error) {
	if !p.IsAdmin() {
		return domain.Settings{}, domain.ErrForbidden
	}
	if err := st.Validate(); err != nil {
		return domain.Settings{}, err
	}
	if err := s.repo.Save(ctx, st); err != nil {
		return domain.Settings{}, err
	}
	return st, nil
}
