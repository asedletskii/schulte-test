package usecase

import (
	"context"

	"schulte/internal/domain"
)

// SubjectInput — данные для создания/изменения испытуемого.
type SubjectInput struct {
	FullName  string
	BirthYear int
	Group     string
	Notes     string
}

// SubjectService — вариант использования «Вести данные об испытуемых».
type SubjectService struct {
	subjects SubjectRepository
	ids      IDGenerator
	clock    Clock
}

// NewSubjectService создаёт сервис управления испытуемыми.
func NewSubjectService(subjects SubjectRepository, ids IDGenerator, clock Clock) *SubjectService {
	return &SubjectService{subjects: subjects, ids: ids, clock: clock}
}

// Create регистрирует нового испытуемого (только администратор).
func (s *SubjectService) Create(ctx context.Context, p Principal, in SubjectInput) (*domain.Subject, error) {
	if !p.IsAdmin() {
		return nil, domain.ErrForbidden
	}
	now := s.clock.Now()
	sub := &domain.Subject{
		ID: s.ids.NewID(), FullName: in.FullName, BirthYear: in.BirthYear,
		Group: in.Group, Notes: in.Notes, CreatedAt: now,
	}
	if err := sub.Validate(now); err != nil {
		return nil, err
	}
	if err := s.subjects.Create(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

// Update изменяет данные испытуемого (только администратор).
func (s *SubjectService) Update(ctx context.Context, p Principal, id string, in SubjectInput) (*domain.Subject, error) {
	if !p.IsAdmin() {
		return nil, domain.ErrForbidden
	}
	sub, err := s.subjects.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	sub.FullName, sub.BirthYear, sub.Group, sub.Notes = in.FullName, in.BirthYear, in.Group, in.Notes
	if err := sub.Validate(s.clock.Now()); err != nil {
		return nil, err
	}
	if err := s.subjects.Update(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

// Delete удаляет испытуемого вместе с его результатами (только администратор).
func (s *SubjectService) Delete(ctx context.Context, p Principal, id string) error {
	if !p.IsAdmin() {
		return domain.ErrForbidden
	}
	return s.subjects.Delete(ctx, id) // каскадное удаление сеансов обеспечивает хранилище
}

// Get возвращает карточку испытуемого с учётом прав доступа.
func (s *SubjectService) Get(ctx context.Context, p Principal, id string) (*domain.Subject, error) {
	if !p.CanAccessSubject(id) {
		return nil, domain.ErrForbidden
	}
	return s.subjects.GetByID(ctx, id)
}

// List возвращает список испытуемых (только администратор).
func (s *SubjectService) List(ctx context.Context, p Principal, group string) ([]domain.Subject, error) {
	if !p.IsAdmin() {
		return nil, domain.ErrForbidden
	}
	return s.subjects.List(ctx, group)
}

// Groups возвращает названия групп (только администратор).
func (s *SubjectService) Groups(ctx context.Context, p Principal) ([]string, error) {
	if !p.IsAdmin() {
		return nil, domain.ErrForbidden
	}
	return s.subjects.Groups(ctx)
}

// ListPublic возвращает краткий список для экрана выбора пользователя
// при входе в режиме тестируемого (аутентификация не требуется).
func (s *SubjectService) ListPublic(ctx context.Context) ([]domain.Subject, error) {
	return s.subjects.List(ctx, "")
}
