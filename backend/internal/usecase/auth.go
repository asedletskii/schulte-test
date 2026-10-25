package usecase

import (
	"context"
	"crypto/subtle"
	"time"

	"schulte/internal/domain"
)

// Session — результат успешной аутентификации.
type Session struct {
	Token     string
	ExpiresAt time.Time
	Principal Principal
}

// AuthService — вариант использования «Войти в систему».
type AuthService struct {
	adminPassword string
	subjects      SubjectRepository
	tokens        TokenService
}

// NewAuthService создаёт сервис аутентификации.
func NewAuthService(adminPassword string, subjects SubjectRepository, tokens TokenService) *AuthService {
	return &AuthService{adminPassword: adminPassword, subjects: subjects, tokens: tokens}
}

// LoginAdmin — вход психолога по паролю (режим администратора).
func (a *AuthService) LoginAdmin(_ context.Context, password string) (Session, error) {
	if subtle.ConstantTimeCompare([]byte(password), []byte(a.adminPassword)) != 1 {
		return Session{}, domain.ErrUnauthorized
	}
	return a.issue(Principal{Role: RoleAdmin})
}

// LoginSubject — вход в режиме тестируемого: выбор своей карточки.
func (a *AuthService) LoginSubject(ctx context.Context, subjectID string) (Session, error) {
	if _, err := a.subjects.GetByID(ctx, subjectID); err != nil {
		return Session{}, err
	}
	return a.issue(Principal{Role: RoleSubject, SubjectID: subjectID})
}

// Authenticate проверяет токен и возвращает пользователя.
func (a *AuthService) Authenticate(token string) (Principal, error) {
	return a.tokens.Parse(token)
}

func (a *AuthService) issue(p Principal) (Session, error) {
	tok, exp, err := a.tokens.Issue(p)
	if err != nil {
		return Session{}, err
	}
	return Session{Token: tok, ExpiresAt: exp, Principal: p}, nil
}
