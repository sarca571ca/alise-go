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
