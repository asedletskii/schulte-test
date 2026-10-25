package postgres_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"schulte/internal/adapter/repository/postgres"
	"schulte/internal/domain"
)

// Интеграционный тест: выполняется только при заданной TEST_DATABASE_URL.
func connect(t *testing.T) *postgres.Store {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL не задан — интеграционные тесты PostgreSQL пропущены")
	}
	st, err := postgres.Connect(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(st.Close)
	if _, err := st.Pool().Exec(context.Background(), `TRUNCATE subjects, settings, test_sessions CASCADE`); err != nil {
		t.Fatal(err)
	}
	return st
}

func TestRepositories(t *testing.T) {
	ctx := context.Background()
	st := connect(t)
	subjects, sessions, settings := st.Subjects(), st.Sessions(), st.Settings()

	now := time.Now().UTC().Truncate(time.Millisecond)
	sub := &domain.Subject{ID: "s1", FullName: "Иван Петров", BirthYear: 2001, Group: "Г1", Notes: "n", CreatedAt: now}
	if err := subjects.Create(ctx, sub); err != nil {
		t.Fatal(err)
	}
	got, err := subjects.GetByID(ctx, "s1")
	if err != nil || got.FullName != "Иван Петров" || !got.CreatedAt.Equal(now) {
		t.Fatalf("GetByID: %v %+v", err, got)
	}
	if _, err := subjects.GetByID(ctx, "нет"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ожидалось ErrNotFound: %v", err)
	}
	sub.FullName = "Иван Сидоров"
	if err := subjects.Update(ctx, sub); err != nil {
		t.Fatal(err)
	}
	if err := subjects.Update(ctx, &domain.Subject{ID: "нет"}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("update несуществующего: %v", err)
	}
	_ = subjects.Create(ctx, &domain.Subject{ID: "s2", FullName: "Анна", Group: "Г2", CreatedAt: now})
	if l, _ := subjects.List(ctx, ""); len(l) != 2 || l[0].FullName != "Анна" {
		t.Errorf("List: %+v", l)
	}
	if l, _ := subjects.List(ctx, "Г2"); len(l) != 1 {
		t.Errorf("List по группе: %+v", l)
	}
	if g, _ := subjects.Groups(ctx); len(g) != 2 || g[0] != "Г1" {
		t.Errorf("Groups: %v", g)
	}

	// сеанс: сохранение, завершение, чтение
	tables := domain.NewTableGenerator(1, 2).GenerateAll()
	ses, err := domain.NewTestSession("t1", "s1", tables, domain.DefaultSettings(), now)
	if err != nil {
		t.Fatal(err)
	}
	if err := sessions.Save(ctx, ses); err != nil {
		t.Fatal(err)
	}
	if err := ses.Complete([]domain.TableResult{{Seconds: 40}, {Seconds: 38}, {Seconds: 36}, {Seconds: 44}, {Seconds: 42}}, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := sessions.Save(ctx, ses); err != nil {
		t.Fatal(err)
	}
	back, err := sessions.GetByID(ctx, "t1")
	if err != nil {
		t.Fatal(err)
	}
	if back.Status != domain.StatusCompleted || back.Metrics == nil || back.Metrics.WorkEfficiency != 40 ||
		len(back.Tables) != 5 || back.Tables[0][0] != tables[0][0] || back.Settings != domain.DefaultSettings() ||
		len(back.Results) != 5 || back.FinishedAt == nil {
		t.Fatalf("сеанс прочитан неверно: %+v", back)
	}
	if err := sessions.Save(ctx, &domain.TestSession{ID: "x", SubjectID: "нет", Status: domain.StatusInProgress,
		Tables: tables, StartedAt: now}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("сеанс несуществующего испытуемого: %v", err)
	}
	if l, _ := sessions.ListBySubject(ctx, "s1"); len(l) != 1 {
		t.Errorf("ListBySubject: %d", len(l))
	}

	// настройки
	if _, err := settings.Get(ctx); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("пустые настройки должны давать ErrNotFound: %v", err)
	}
	want := domain.Settings{FontSize: 40, BackgroundColor: "#eeeeee", ShowTimer: false, SoundOnError: true}
	if err := settings.Save(ctx, want); err != nil {
		t.Fatal(err)
	}
	if err := settings.Save(ctx, want); err != nil { // upsert
		t.Fatal(err)
	}
	if got, _ := settings.Get(ctx); got != want {
		t.Errorf("настройки: %+v", got)
	}

	// каскадное удаление
	if err := subjects.Delete(ctx, "s1"); err != nil {
		t.Fatal(err)
	}
	if _, err := sessions.GetByID(ctx, "t1"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("сеанс должен быть удалён каскадно: %v", err)
	}
	if err := subjects.Delete(ctx, "s1"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("повторное удаление: %v", err)
	}
}
