// Пакет usecase реализует прикладные сценарии (варианты использования).
// Он зависит только от пакета domain и описывает порты (интерфейсы),
// которые реализуются внешними слоями (БД, токены, часы и т. д.).
package usecase

import (
	"context"
	"time"

	"schulte/internal/domain"
)

// Role — роль пользователя системы.
type Role string

const (
	RoleAdmin   Role = "admin"   // психолог (режим администратора)
	RoleSubject Role = "subject" // испытуемый (режим тестируемого)
)

// Principal — аутентифицированный пользователь.
type Principal struct {
	Role      Role
	SubjectID string // заполнен для RoleSubject
}

// IsAdmin сообщает, обладает ли пользователь правами администратора.
func (p Principal) IsAdmin() bool { return p.Role == RoleAdmin }

// CanAccessSubject — доступ к данным испытуемого: админ — ко всем, испытуемый — только к своим.
func (p Principal) CanAccessSubject(subjectID string) bool {
	return p.IsAdmin() || (p.Role == RoleSubject && p.SubjectID == subjectID)
}

// ---- Порты хранилищ ----

// SubjectRepository — хранилище испытуемых.
type SubjectRepository interface {
	Create(ctx context.Context, s *domain.Subject) error
	Update(ctx context.Context, s *domain.Subject) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*domain.Subject, error)
	// List возвращает испытуемых; при непустом group — только указанной группы.
	List(ctx context.Context, group string) ([]domain.Subject, error)
	Groups(ctx context.Context) ([]string, error)
}

// SessionRepository — хранилище сеансов тестирования.
type SessionRepository interface {
	// Save создаёт или обновляет сеанс.
	Save(ctx context.Context, s *domain.TestSession) error
	GetByID(ctx context.Context, id string) (*domain.TestSession, error)
	ListBySubject(ctx context.Context, subjectID string) ([]*domain.TestSession, error)
	Delete(ctx context.Context, id string) error
}

// SettingsRepository — хранилище глобальных параметров тестирования.
type SettingsRepository interface {
	Get(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// ---- Вспомогательные порты ----

// TokenService выпускает и проверяет токены доступа.
type TokenService interface {
	Issue(p Principal) (string, time.Time, error)
	Parse(token string) (Principal, error)
}

// IDGenerator формирует уникальные идентификаторы.
type IDGenerator interface{ NewID() string }

// Clock возвращает текущее время (подменяется в тестах).
type Clock interface{ Now() time.Time }

// TablesSource поставляет таблицы Шульте для нового теста.
type TablesSource interface{ GenerateAll() [][]int }
