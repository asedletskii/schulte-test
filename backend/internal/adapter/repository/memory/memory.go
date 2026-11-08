package memory

import (
	"context"
	"sort"
	"sync"

	"schulte/internal/domain"
)

// содержит все три репозитория и общую блокировку
type Store struct {
	mu       sync.RWMutex
	subjects map[string]domain.Subject
	sessions map[string]domain.TestSession
	settings *domain.Settings
}

// создаёт пустое хранилище.
func NewStore() *Store {
	return &Store{subjects: map[string]domain.Subject{}, sessions: map[string]domain.TestSession{}}
}

// ---- SubjectRepository ----

type SubjectRepo struct{ s *Store }

func (s *Store) Subjects() *SubjectRepo { return &SubjectRepo{s} }

func (r *SubjectRepo) Create(_ context.Context, sub *domain.Subject) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.subjects[sub.ID] = *sub
	return nil
}

func (r *SubjectRepo) Update(_ context.Context, sub *domain.Subject) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.subjects[sub.ID]; !ok {
		return domain.ErrNotFound
	}
	r.s.subjects[sub.ID] = *sub
	return nil
}

func (r *SubjectRepo) Delete(_ context.Context, id string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.subjects[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.s.subjects, id)
	for sid, ses := range r.s.sessions {
		if ses.SubjectID == id {
			delete(r.s.sessions, sid)
		}
	}
	return nil
}

func (r *SubjectRepo) GetByID(_ context.Context, id string) (*domain.Subject, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	sub, ok := r.s.subjects[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &sub, nil
}

func (r *SubjectRepo) List(_ context.Context, group string) ([]domain.Subject, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]domain.Subject, 0, len(r.s.subjects))
	for _, sub := range r.s.subjects {
		if group == "" || sub.Group == group {
			out = append(out, sub)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].FullName < out[j].FullName })
	return out, nil
}

func (r *SubjectRepo) Groups(_ context.Context) ([]string, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	set := map[string]struct{}{}
	for _, sub := range r.s.subjects {
		if sub.Group != "" {
			set[sub.Group] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for g := range set {
		out = append(out, g)
	}
	sort.Strings(out)
	return out, nil
}

// ---- SessionRepository ----

type SessionRepo struct{ s *Store }

func (s *Store) Sessions() *SessionRepo { return &SessionRepo{s} }

func (r *SessionRepo) Save(_ context.Context, ses *domain.TestSession) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.subjects[ses.SubjectID]; !ok {
		return domain.ErrNotFound
	}
	r.s.sessions[ses.ID] = clone(*ses)
	return nil
}

func (r *SessionRepo) GetByID(_ context.Context, id string) (*domain.TestSession, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	ses, ok := r.s.sessions[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	c := clone(ses)
	return &c, nil
}

func (r *SessionRepo) ListBySubject(_ context.Context, subjectID string) ([]*domain.TestSession, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []*domain.TestSession{}
	for _, ses := range r.s.sessions {
		if ses.SubjectID == subjectID {
			c := clone(ses)
			out = append(out, &c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartedAt.After(out[j].StartedAt) })
	return out, nil
}

func (r *SessionRepo) Delete(_ context.Context, id string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.sessions[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.s.sessions, id)
	return nil
}

func clone(s domain.TestSession) domain.TestSession {
	c := s
	c.Tables = make([][]int, len(s.Tables))
	for i, t := range s.Tables {
		c.Tables[i] = append([]int(nil), t...)
	}
	c.Results = append([]domain.TableResult(nil), s.Results...)
	if s.Metrics != nil {
		m := *s.Metrics
		m.FatigueCurve = append([]float64(nil), s.Metrics.FatigueCurve...)
		c.Metrics = &m
	}
	if s.FinishedAt != nil {
		t := *s.FinishedAt
		c.FinishedAt = &t
	}
	return c
}

// ---- SettingsRepository ----

type SettingsRepo struct{ s *Store }

func (s *Store) Settings() *SettingsRepo { return &SettingsRepo{s} }

func (r *SettingsRepo) Get(_ context.Context) (domain.Settings, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	if r.s.settings == nil {
		return domain.Settings{}, domain.ErrNotFound
	}
	return *r.s.settings, nil
}

func (r *SettingsRepo) Save(_ context.Context, st domain.Settings) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.settings = &st
	return nil
}
