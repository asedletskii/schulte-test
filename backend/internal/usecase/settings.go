package usecase

import (
	"context"
	"errors"

	"schulte/internal/domain"
)

// вариант использования «Настроить параметры тестирования»
type SettingsService struct {
	repo SettingsRepository
}

func NewSettingsService(repo SettingsRepository) *SettingsService {
	return &SettingsService{repo: repo}
}

func (s *SettingsService) Get(ctx context.Context) (domain.Settings, error) {
	st, err := s.repo.Get(ctx)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.DefaultSettings(), nil
	}
	if err != nil {
		return domain.Settings{}, err
	}
	return st, nil
}

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
