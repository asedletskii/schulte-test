// Пакет postgres — реализация портов хранилищ на PostgreSQL (драйвер pgx).
package postgres

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"schulte/internal/domain"
)

//go:embed schema.sql
var schemaSQL string

// Store — подключение к БД; создаёт репозитории.
type Store struct{ pool *pgxpool.Pool }

// Connect открывает пул соединений и применяет схему БД.
func Connect(ctx context.Context, url string) (*Store, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("подключение к БД: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("БД недоступна: %w", err)
	}
	if _, err := pool.Exec(ctx, schemaSQL); err != nil {
		pool.Close()
		return nil, fmt.Errorf("применение схемы: %w", err)
	}
	return &Store{pool: pool}, nil
}

// Close закрывает пул соединений.
func (s *Store) Close() { s.pool.Close() }

// Pool — доступ к пулу (используется в интеграционных тестах).
func (s *Store) Pool() *pgxpool.Pool { return s.pool }

func (s *Store) Subjects() *SubjectRepo  { return &SubjectRepo{s.pool} }
func (s *Store) Sessions() *SessionRepo  { return &SessionRepo{s.pool} }
func (s *Store) Settings() *SettingsRepo { return &SettingsRepo{s.pool} }

func mapErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "23503" { // foreign_key_violation
		return domain.ErrNotFound
	}
	return err
}

// ---- Испытуемые ----

type SubjectRepo struct{ pool *pgxpool.Pool }

const subjectCols = `id, full_name, birth_year, group_name, notes, created_at`

func scanSubject(row pgx.Row) (*domain.Subject, error) {
	var s domain.Subject
	if err := row.Scan(&s.ID, &s.FullName, &s.BirthYear, &s.Group, &s.Notes, &s.CreatedAt); err != nil {
		return nil, mapErr(err)
	}
	return &s, nil
}

func (r *SubjectRepo) Create(ctx context.Context, s *domain.Subject) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO subjects (`+subjectCols+`) VALUES ($1,$2,$3,$4,$5,$6)`,
		s.ID, s.FullName, s.BirthYear, s.Group, s.Notes, s.CreatedAt)
	return err
}

func (r *SubjectRepo) Update(ctx context.Context, s *domain.Subject) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE subjects SET full_name=$2, birth_year=$3, group_name=$4, notes=$5 WHERE id=$1`,
		s.ID, s.FullName, s.BirthYear, s.Group, s.Notes)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SubjectRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM subjects WHERE id=$1`, id) // результаты удаляются каскадно
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SubjectRepo) GetByID(ctx context.Context, id string) (*domain.Subject, error) {
	return scanSubject(r.pool.QueryRow(ctx, `SELECT `+subjectCols+` FROM subjects WHERE id=$1`, id))
}

func (r *SubjectRepo) List(ctx context.Context, group string) ([]domain.Subject, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+subjectCols+` FROM subjects WHERE ($1 = '' OR group_name = $1) ORDER BY full_name`, group)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Subject{}
	for rows.Next() {
		s, err := scanSubject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *s)
	}
	return out, rows.Err()
}

func (r *SubjectRepo) Groups(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT DISTINCT group_name FROM subjects WHERE group_name <> '' ORDER BY group_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var g string
		if err := rows.Scan(&g); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

// ---- Сеансы тестирования ----

type SessionRepo struct{ pool *pgxpool.Pool }

const sessionCols = `id, subject_id, status, tables, settings, started_at, finished_at, results, metrics`

// Save выполняет upsert сеанса.
func (r *SessionRepo) Save(ctx context.Context, s *domain.TestSession) error {
	tables, _ := json.Marshal(s.Tables)
	settings, _ := json.Marshal(settingsDTO{s.Settings.FontSize, s.Settings.BackgroundColor, s.Settings.ShowTimer, s.Settings.SoundOnError})
	var results, metrics []byte
	if s.Results != nil {
		results, _ = json.Marshal(s.Results)
	}
	if s.Metrics != nil {
		metrics, _ = json.Marshal(s.Metrics)
	}
	_, err := r.pool.Exec(ctx, `
INSERT INTO test_sessions (`+sessionCols+`) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
ON CONFLICT (id) DO UPDATE SET
    status=EXCLUDED.status, finished_at=EXCLUDED.finished_at,
    results=EXCLUDED.results, metrics=EXCLUDED.metrics`,
		s.ID, s.SubjectID, string(s.Status), tables, settings, s.StartedAt, s.FinishedAt, results, metrics)
	return mapErr(err)
}

type settingsDTO struct {
	FontSize        int    `json:"fontSize"`
	BackgroundColor string `json:"backgroundColor"`
	ShowTimer       bool   `json:"showTimer"`
	SoundOnError    bool   `json:"soundOnError"`
}

func scanSession(row pgx.Row) (*domain.TestSession, error) {
	var (
		s                domain.TestSession
		status           string
		tables, settings []byte
		results, metrics []byte
		finished         *time.Time
	)
	if err := row.Scan(&s.ID, &s.SubjectID, &status, &tables, &settings, &s.StartedAt, &finished, &results, &metrics); err != nil {
		return nil, mapErr(err)
	}
	s.Status = domain.SessionStatus(status)
	s.FinishedAt = finished
	if err := json.Unmarshal(tables, &s.Tables); err != nil {
		return nil, err
	}
	var st settingsDTO
	if err := json.Unmarshal(settings, &st); err != nil {
		return nil, err
	}
	s.Settings = domain.Settings{FontSize: st.FontSize, BackgroundColor: st.BackgroundColor, ShowTimer: st.ShowTimer, SoundOnError: st.SoundOnError}
	if results != nil {
		if err := json.Unmarshal(results, &s.Results); err != nil {
			return nil, err
		}
	}
	if metrics != nil {
		var m domain.Metrics
		if err := json.Unmarshal(metrics, &m); err != nil {
			return nil, err
		}
		s.Metrics = &m
	}
	return &s, nil
}

func (r *SessionRepo) GetByID(ctx context.Context, id string) (*domain.TestSession, error) {
	return scanSession(r.pool.QueryRow(ctx, `SELECT `+sessionCols+` FROM test_sessions WHERE id=$1`, id))
}

func (r *SessionRepo) ListBySubject(ctx context.Context, subjectID string) ([]*domain.TestSession, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+sessionCols+` FROM test_sessions WHERE subject_id=$1 ORDER BY started_at DESC`, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*domain.TestSession{}
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *SessionRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM test_sessions WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ---- Настройки ----

type SettingsRepo struct{ pool *pgxpool.Pool }

func (r *SettingsRepo) Get(ctx context.Context) (domain.Settings, error) {
	var s domain.Settings
	err := r.pool.QueryRow(ctx,
		`SELECT font_size, background_color, show_timer, sound_on_error FROM settings WHERE id=1`).
		Scan(&s.FontSize, &s.BackgroundColor, &s.ShowTimer, &s.SoundOnError)
	if err != nil {
		return domain.Settings{}, mapErr(err)
	}
	return s, nil
}

func (r *SettingsRepo) Save(ctx context.Context, s domain.Settings) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO settings (id, font_size, background_color, show_timer, sound_on_error) VALUES (1,$1,$2,$3,$4)
ON CONFLICT (id) DO UPDATE SET font_size=EXCLUDED.font_size, background_color=EXCLUDED.background_color,
    show_timer=EXCLUDED.show_timer, sound_on_error=EXCLUDED.sound_on_error`,
		s.FontSize, s.BackgroundColor, s.ShowTimer, s.SoundOnError)
	return err
}
