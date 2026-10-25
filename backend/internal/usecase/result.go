package usecase

import (
	"context"

	"schulte/internal/domain"
)

// ResultService — варианты использования «Просмотреть результаты»,
// «Просмотреть динамику результатов» и «Сравнить результаты».
type ResultService struct {
	subjects SubjectRepository
	sessions SessionRepository
}

// NewResultService создаёт сервис работы с результатами.
func NewResultService(subjects SubjectRepository, sessions SessionRepository) *ResultService {
	return &ResultService{subjects: subjects, sessions: sessions}
}

// List возвращает завершённые тестирования испытуемого (новые — первыми).
func (r *ResultService) List(ctx context.Context, p Principal, subjectID string) ([]*domain.TestSession, error) {
	if !p.CanAccessSubject(subjectID) {
		return nil, domain.ErrForbidden
	}
	all, err := r.sessions.ListBySubject(ctx, subjectID)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.TestSession, 0, len(all))
	for _, s := range all {
		if s.Status == domain.StatusCompleted {
			out = append(out, s)
		}
	}
	return out, nil
}

// Get возвращает один результат.
func (r *ResultService) Get(ctx context.Context, p Principal, id string) (*domain.TestSession, error) {
	s, err := r.sessions.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !p.CanAccessSubject(s.SubjectID) {
		return nil, domain.ErrForbidden
	}
	return s, nil
}

// Delete удаляет результат (только администратор).
func (r *ResultService) Delete(ctx context.Context, p Principal, id string) error {
	if !p.IsAdmin() {
		return domain.ErrForbidden
	}
	return r.sessions.Delete(ctx, id)
}

// Dynamics формирует отчёт о динамике результатов испытуемого.
func (r *ResultService) Dynamics(ctx context.Context, p Principal, subjectID string) (domain.Dynamics, error) {
	if !p.CanAccessSubject(subjectID) {
		return domain.Dynamics{}, domain.ErrForbidden
	}
	all, err := r.sessions.ListBySubject(ctx, subjectID)
	if err != nil {
		return domain.Dynamics{}, err
	}
	return domain.BuildDynamics(all), nil
}

// Compare сравнивает два тестирования (только администратор): одного или разных испытуемых.
func (r *ResultService) Compare(ctx context.Context, p Principal, idA, idB string) (domain.Comparison, error) {
	if !p.IsAdmin() {
		return domain.Comparison{}, domain.ErrForbidden
	}
	a, err := r.sessions.GetByID(ctx, idA)
	if err != nil {
		return domain.Comparison{}, err
	}
	b, err := r.sessions.GetByID(ctx, idB)
	if err != nil {
		return domain.Comparison{}, err
	}
	return domain.Compare(a, b)
}

// GroupStats формирует сводку по группе испытуемых (только администратор).
func (r *ResultService) GroupStats(ctx context.Context, p Principal, group string) (domain.GroupStats, error) {
	if !p.IsAdmin() {
		return domain.GroupStats{}, domain.ErrForbidden
	}
	subs, err := r.subjects.List(ctx, group)
	if err != nil {
		return domain.GroupStats{}, err
	}
	bySubject := make(map[string][]*domain.TestSession, len(subs))
	for _, s := range subs {
		list, err := r.sessions.ListBySubject(ctx, s.ID)
		if err != nil {
			return domain.GroupStats{}, err
		}
		bySubject[s.ID] = list
	}
	return domain.BuildGroupStats(group, subs, bySubject), nil
}
