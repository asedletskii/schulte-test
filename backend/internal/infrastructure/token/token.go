// Пакет token реализует выпуск и проверку подписанных токенов доступа
// (HMAC-SHA256) без внешних зависимостей.
package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"schulte/internal/domain"
	"schulte/internal/usecase"
)

type claims struct {
	Role usecase.Role `json:"role"`
	Sub  string       `json:"sub,omitempty"`
	Exp  int64        `json:"exp"`
}

// Service — реализация usecase.TokenService.
type Service struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

// New создаёт сервис токенов с заданным секретом и временем жизни.
func New(secret string, ttl time.Duration) *Service {
	return &Service{secret: []byte(secret), ttl: ttl, now: time.Now}
}

// WithClock подменяет источник времени (для тестов).
func (s *Service) WithClock(now func() time.Time) *Service { s.now = now; return s }

var enc = base64.RawURLEncoding

// Issue выпускает токен вида base64url(payload).base64url(hmac).
func (s *Service) Issue(p usecase.Principal) (string, time.Time, error) {
	exp := s.now().Add(s.ttl)
	payload, err := json.Marshal(claims{Role: p.Role, Sub: p.SubjectID, Exp: exp.Unix()})
	if err != nil {
		return "", time.Time{}, err
	}
	body := enc.EncodeToString(payload)
	return body + "." + s.sign(body), exp, nil
}

// Parse проверяет подпись и срок действия.
func (s *Service) Parse(tok string) (usecase.Principal, error) {
	body, sig, ok := strings.Cut(tok, ".")
	if !ok || !hmac.Equal([]byte(sig), []byte(s.sign(body))) {
		return usecase.Principal{}, domain.ErrUnauthorized
	}
	raw, err := enc.DecodeString(body)
	if err != nil {
		return usecase.Principal{}, domain.ErrUnauthorized
	}
	var c claims
	if err := json.Unmarshal(raw, &c); err != nil || s.now().Unix() >= c.Exp {
		return usecase.Principal{}, domain.ErrUnauthorized
	}
	if c.Role != usecase.RoleAdmin && c.Role != usecase.RoleSubject {
		return usecase.Principal{}, domain.ErrUnauthorized
	}
	return usecase.Principal{Role: c.Role, SubjectID: c.Sub}, nil
}

func (s *Service) sign(body string) string {
	m := hmac.New(sha256.New, s.secret)
	m.Write([]byte(body))
	return enc.EncodeToString(m.Sum(nil))
}
