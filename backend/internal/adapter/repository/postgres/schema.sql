CREATE TABLE IF NOT EXISTS subjects (
    id          TEXT PRIMARY KEY,
    full_name   TEXT        NOT NULL,
    birth_year  INTEGER     NOT NULL DEFAULT 0,
    group_name  TEXT        NOT NULL DEFAULT '',
    notes       TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_subjects_group ON subjects (group_name);

CREATE TABLE IF NOT EXISTS settings (
    id               SMALLINT PRIMARY KEY CHECK (id = 1),
    font_size        INTEGER NOT NULL,
    background_color TEXT    NOT NULL,
    show_timer       BOOLEAN NOT NULL,
    sound_on_error   BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS test_sessions (
    id           TEXT PRIMARY KEY,
    subject_id   TEXT        NOT NULL REFERENCES subjects (id) ON DELETE CASCADE,
    status       TEXT        NOT NULL CHECK (status IN ('in_progress', 'completed')),
    tables       JSONB       NOT NULL,
    settings     JSONB       NOT NULL,
    started_at   TIMESTAMPTZ NOT NULL,
    finished_at  TIMESTAMPTZ,
    results      JSONB,
    metrics      JSONB
);

CREATE INDEX IF NOT EXISTS idx_sessions_subject ON test_sessions (subject_id, started_at DESC);
