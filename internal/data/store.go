package data

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

const schema = `
CREATE TABLE IF NOT EXISTS hnm_timers (
    id            TEXT PRIMARY KEY,
    guild_id      TEXT NOT NULL,
    channel_id    TEXT NOT NULL,
    hnm_id        TEXT NOT NULL,
    last_kill     TEXT NOT NULL,
    days_since_hq INTEGER NOT NULL,
	is_notified   INTEGER NOT NULL DEFAULT 0,
    created_at    TEXT NOT NULL,
    updated_at    TEXT NOT NULL,
    UNIQUE (guild_id, channel_id, hnm_id)
);

CREATE TABLE IF NOT EXISTS hnm_timer_boards (
    guild_id   TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    PRIMARY KEY (guild_id, channel_id)
);

CREATE TABLE IF NOT EXISTS hnm_camp_channels (
    id              TEXT PRIMARY KEY,
    guild_id        TEXT NOT NULL,
    channel_id      TEXT NOT NULL,
    hnm_id          TEXT NOT NULL,
    last_kill       TEXT NOT NULL,
    days_since_hq   INTEGER NOT NULL,
    seq             INTEGER NOT NULL,
    is_closed       INTEGER NOT NULL DEFAULT 0,
    is_enraged      INTEGER NOT NULL DEFAULT 0,
    last_window_idx INTEGER NOT NULL DEFAULT 0,
    created_at      TEXT NOT NULL,
    updated_at      TEXT NOT NULL,
	move_scheduled  INTEGER NOT NULL DEFAULT 0,
    UNIQUE (guild_id, hnm_id, last_kill, days_since_hq)
);
`

func NewStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return &Store{DB: db}, nil
}
