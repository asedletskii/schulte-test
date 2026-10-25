package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"schulte/internal/adapter/repository/memory"
	"schulte/internal/domain"
	"schulte/internal/infrastructure/token"
	"schulte/internal/usecase"
)

// ---- тестовые двойники ----

type seqIDs struct{ n int }

func (s *seqIDs) NewID() string { s.n++; return fmt.Sprintf("id-%d", s.n) }

type fixedClock struct{ t time.Time }

func (c *fixedClock) Now() time.Time { c.t = c.t.Add(time.Minute); return c.t }

type env struct {
	subjects *usecase.SubjectService
	tests    *usecase.TestService
	results  *usecase.ResultService
	settings *usecase.SettingsService
	auth     *usecase.AuthService
}

func setup() env {
	st := memory.NewStore()
	ids, clock := &seqIDs{}, &fixedClock{t: time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)}
	settings := usecase.NewSettingsService(st.Settings())
	return env{
		subjects: usecase.NewSubjectService(st.Subjects(), ids, clock),
		tests:    usecase.NewTestService(st.Subjects(), st.Sessions(), settings, domain.NewTableGenerator(7, 7), ids, clock),
		results:  usecase.NewResultService(st.Subjects(), st.Sessions()),
		settings: settings,
		auth:     usecase.NewAuthService("secret", st.Subjects(), token.New("k", time.Hour)),
	}
}

var (
	admin = usecase.Principal{Role: usecase.RoleAdmin}
	ctx   = context.Background()
)

func subjectP(id string) usecase.Principal {
	return usecase.Principal{Role: usecase.RoleSubject, SubjectID: id}
}

func secs(v float64) []domain.TableResult {
	out := make([]domain.TableResult, domain.TablesCount)
	for i := range out {
		out[i] = domain.TableResult{Number: i + 1, Seconds: v}
	}
	return out
}

// ---- Вариант использования: Войти в систему ----

func TestAuth(t *testing.T) {
	e := setup()
	if _, err := e.auth.LoginAdmin(ctx, "wrong"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("неверный пароль должен отклоняться: %v", err)
	}
	s, err := e.auth.LoginAdmin(ctx, "secret")
	if err != nil || s.Principal.Role != usecase.RoleAdmin {
		t.Fatalf("вход администратора: %v %+v", err, s)
	}
	p, err := e.auth.Authenticate(s.Token)
	if err != nil || !p.IsAdmin() {
		t.Fatalf("токен не принят: %v", err)
	}
	if _, err := e.auth.LoginSubject(ctx, "нет-такого"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("вход несуществующего испытуемого: %v", err)
	}
	sub, _ := e.subjects.Create(ctx, admin, usecase.SubjectInput{FullName: "Иван Петров"})
	ss, err := e.auth.LoginSubject(ctx, sub.ID)
	if err != nil || ss.Principal.SubjectID != sub.ID {
		t.Fatalf("вход испытуемого: %v", err)
	}
}

// ---- Вариант использования: Вести данные об испытуемых ----

func TestSubjectCRUDAndPermissions(t *testing.T) {
	e := setup()
	if _, err := e.subjects.Create(ctx, subjectP("x"), usecase.SubjectInput{FullName: "Иван"}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("испытуемый не может создавать испытуемых")
	}
	if _, err := e.subjects.Create(ctx, admin, usecase.SubjectInput{FullName: ""}); !errors.Is(err, domain.ErrValidation) {
		t.Fatal("пустое ФИО должно отклоняться")
	}
	a, err := e.subjects.Create(ctx, admin, usecase.SubjectInput{FullName: "Анна Смирнова", Group: "Г1"})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := e.subjects.Create(ctx, admin, usecase.SubjectInput{FullName: "Борис Орлов", Group: "Г2"})

	list, _ := e.subjects.List(ctx, admin, "Г1")
	if len(list) != 1 || list[0].ID != a.ID {
		t.Fatalf("фильтр по группе: %+v", list)
	}
	groups, _ := e.subjects.Groups(ctx, admin)
	if len(groups) != 2 {
		t.Errorf("группы: %v", groups)
	}
	upd, err := e.subjects.Update(ctx, admin, a.ID, usecase.SubjectInput{FullName: "Анна Смирнова-Иванова", Group: "Г1"})
	if err != nil || upd.FullName != "Анна Смирнова-Иванова" {
		t.Fatalf("обновление: %v", err)
	}
	if _, err := e.subjects.Get(ctx, subjectP(a.ID), b.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Error("испытуемый не должен видеть чужую карточку")
	}
	if err := e.subjects.Delete(ctx, admin, b.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := e.subjects.Get(ctx, admin, b.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Error("удалённый испытуемый должен быть недоступен")
	}
}

// ---- Вариант использования: Пройти тестирование ----

func TestRunTestAndViewResult(t *testing.T) {
	e := setup()
	sub, _ := e.subjects.Create(ctx, admin, usecase.SubjectInput{FullName: "Иван Петров"})
	me := subjectP(sub.ID)

	session, err := e.tests.Start(ctx, me, sub.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(session.Tables) != domain.TablesCount || session.Status != domain.StatusInProgress {
		t.Fatalf("сеанс начат неверно: %+v", session)
	}
	if session.Settings != domain.DefaultSettings() {
		t.Error("должны применяться настройки по умолчанию")
	}
	// нельзя завершить чужой тест
	other, _ := e.subjects.Create(ctx, admin, usecase.SubjectInput{FullName: "Пётр Иванов"})
	if _, err := e.tests.Finish(ctx, subjectP(other.ID), session.ID, secs(30)); !errors.Is(err, domain.ErrForbidden) {
		t.Error("чужой тест нельзя завершать")
	}
	done, err := e.tests.Finish(ctx, me, session.ID, secs(30))
	if err != nil {
		t.Fatal(err)
	}
	if done.Metrics == nil || done.Metrics.WorkEfficiency != 30 || done.Metrics.WarmUp != 1 {
		t.Fatalf("показатели: %+v", done.Metrics)
	}
	if _, err := e.tests.Finish(ctx, me, session.ID, secs(30)); !errors.Is(err, domain.ErrConflict) {
		t.Error("повторное завершение должно давать конфликт")
	}
	list, _ := e.results.List(ctx, me, sub.ID)
	if len(list) != 1 {
		t.Fatalf("в списке результатов должен быть один тест: %d", len(list))
	}
	if _, err := e.results.List(ctx, subjectP(other.ID), sub.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Error("чужие результаты недоступны")
	}
}

func TestStartTestWithCustomSettings(t *testing.T) {
	e := setup()
	sub, _ := e.subjects.Create(ctx, admin, usecase.SubjectInput{FullName: "Иван Петров"})
	custom := domain.Settings{FontSize: 40, BackgroundColor: "#ffeecc", ShowTimer: false, SoundOnError: false}
	s, err := e.tests.Start(ctx, admin, sub.ID, &custom)
	if err != nil || s.Settings != custom {
		t.Fatalf("пользовательские настройки не применены: %v %+v", err, s.Settings)
	}
	bad := domain.Settings{FontSize: 1, BackgroundColor: "#fff"}
	if _, err := e.tests.Start(ctx, admin, sub.ID, &bad); !errors.Is(err, domain.ErrValidation) {
		t.Error("некорректные настройки должны отклоняться")
	}
	if _, err := e.tests.Start(ctx, admin, "нет", nil); !errors.Is(err, domain.ErrNotFound) {
		t.Error("тест для несуществующего испытуемого")
	}
}

// ---- Вариант использования: Настроить параметры ----

func TestSettings(t *testing.T) {
	e := setup()
	got, err := e.settings.Get(ctx)
	if err != nil || got != domain.DefaultSettings() {
		t.Fatalf("по умолчанию: %v %+v", err, got)
	}
	want := domain.Settings{FontSize: 36, BackgroundColor: "#eeeeee", ShowTimer: false, SoundOnError: true}
	if _, err := e.settings.Update(ctx, subjectP("x"), want); !errors.Is(err, domain.ErrForbidden) {
		t.Error("испытуемый не может менять глобальные настройки")
	}
	if _, err := e.settings.Update(ctx, admin, want); err != nil {
		t.Fatal(err)
	}
	if got, _ := e.settings.Get(ctx); got != want {
		t.Errorf("настройки не сохранены: %+v", got)
	}
	if _, err := e.settings.Update(ctx, admin, domain.Settings{FontSize: 1}); !errors.Is(err, domain.ErrValidation) {
		t.Error("некорректные настройки должны отклоняться")
	}
}

// ---- Динамика, сравнение, группа ----

func TestDynamicsCompareGroup(t *testing.T) {
	e := setup()
	a, _ := e.subjects.Create(ctx, admin, usecase.SubjectInput{FullName: "Анна", Group: "Г"})
	b, _ := e.subjects.Create(ctx, admin, usecase.SubjectInput{FullName: "Борис", Group: "Г"})

	run := func(sub *domain.Subject, v float64) string {
		s, err := e.tests.Start(ctx, admin, sub.ID, nil)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := e.tests.Finish(ctx, admin, s.ID, secs(v)); err != nil {
			t.Fatal(err)
		}
		return s.ID
	}
	a1 := run(a, 50)
	run(a, 45)
	run(a, 40)
	b1 := run(b, 35)

	d, err := e.results.Dynamics(ctx, subjectP(a.ID), a.ID)
	if err != nil || len(d.Points) != 3 || d.Trend != domain.TrendImproving {
		t.Fatalf("динамика: %v %+v", err, d)
	}
	if _, err := e.results.Dynamics(ctx, subjectP(b.ID), a.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Error("динамика чужого испытуемого недоступна")
	}
	c, err := e.results.Compare(ctx, admin, a1, b1)
	if err != nil || c.Better != "B" {
		t.Fatalf("сравнение: %v %+v", err, c)
	}
	if _, err := e.results.Compare(ctx, subjectP(a.ID), a1, b1); !errors.Is(err, domain.ErrForbidden) {
		t.Error("сравнение доступно только администратору")
	}
	g, err := e.results.GroupStats(ctx, admin, "Г")
	if err != nil || g.ParticipantsCnt != 2 {
		t.Fatalf("группа: %v %+v", err, g)
	}
	// последний тест Анны — 40 с, Бориса — 35 с → среднее 37,5; Борис первый
	if g.AvgEfficiency != 37.5 || g.Members[1].Rank != 1 {
		t.Errorf("статистика группы: %+v", g)
	}
	// удаление испытуемого каскадно удаляет результаты
	if err := e.subjects.Delete(ctx, admin, a.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := e.results.Get(ctx, admin, a1); !errors.Is(err, domain.ErrNotFound) {
		t.Error("результаты удалённого испытуемого должны исчезнуть")
	}
}
