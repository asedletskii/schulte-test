package token

import (
	"errors"
	"testing"
	"time"

	"schulte/internal/domain"
	"schulte/internal/usecase"
)

func TestIssueParse(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s := New("secret", time.Hour).WithClock(func() time.Time { return now })

	tok, _, err := s.Issue(usecase.Principal{Role: usecase.RoleSubject, SubjectID: "u1"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := s.Parse(tok)
	if err != nil || p.Role != usecase.RoleSubject || p.SubjectID != "u1" {
		t.Fatalf("разбор токена: %v %+v", err, p)
	}

	// истёк срок действия
	later := New("secret", time.Hour).WithClock(func() time.Time { return now.Add(2 * time.Hour) })
	if _, err := later.Parse(tok); !errors.Is(err, domain.ErrUnauthorized) {
		t.Error("просроченный токен должен отклоняться")
	}
	// другой секрет
	if _, err := New("other", time.Hour).WithClock(func() time.Time { return now }).Parse(tok); !errors.Is(err, domain.ErrUnauthorized) {
		t.Error("токен с чужой подписью должен отклоняться")
	}
	// подделка полезной нагрузки
	for _, bad := range []string{"", "abc", "a.b", tok + "x", "!!!." + tok} {
		if _, err := s.Parse(bad); !errors.Is(err, domain.ErrUnauthorized) {
			t.Errorf("токен %q должен отклоняться", bad)
		}
	}
}
