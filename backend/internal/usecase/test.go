package usecase

import (
	"context"

	"schulte/internal/domain"
)

// TestService — вариант использования «Пройти тестирование по методике Шульте».
type TestService struct {
	subjects SubjectRepository
	sessions SessionRepository
	settings *SettingsService
	tables   TablesSource
	ids      IDGenerator
	clock    Clock
}

// NewTestService создаёт сервис проведения тестирования.
func NewTestService(subjects SubjectRepository, sessions SessionRepository, settings *SettingsService,
	tables TablesSource, ids IDGenerator, clock Clock) *TestService {
	return &TestService{subjects: subjects, sessions: sessions, settings: settings, tables: tables, ids: ids, clock: clock}
}

// Start начинает тестирование: формирует пять таблиц и фиксирует параметры.
// override — необязательные визуальные параметры, выбранные перед стартом.
func (t *TestService) Start(ctx context.Context, p Principal, subjectID string, override *domain.Settings) (*domain.TestSession, error) {
	if !p.CanAccessSubject(subjectID) {
		return nil, domain.ErrForbidden
	}
	if _, err := t.subjects.GetByID(ctx, subjectID); err != nil {
		return nil, err
	}
	st, err := t.settings.Get(ctx)
	if err != nil {
		return nil, err
	}
	if override != nil {
		st = *override
	}
	session, err := domain.NewTestSession(t.ids.NewID(), subjectID, t.tables.GenerateAll(), st, t.clock.Now())
	if err != nil {
		return nil, err
	}
	if err := t.sessions.Save(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

// Finish принимает результаты прохождения таблиц, рассчитывает показатели и сохраняет сеанс.
func (t *TestService) Finish(ctx context.Context, p Principal, sessionID string, results []domain.TableResult) (*domain.TestSession, error) {
	session, err := t.sessions.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if !p.CanAccessSubject(session.SubjectID) {
		return nil, domain.ErrForbidden
	}
	if err := session.Complete(results, t.clock.Now()); err != nil {
		return nil, err
	}
	if err := t.sessions.Save(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}
