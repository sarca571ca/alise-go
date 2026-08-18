package data

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

const schema = `
CREATE TABLE IF NOT EXISTS hnm_timers (
    id            	TEXT PRIMARY KEY,
    guild_id      	TEXT NOT NULL,
    channel_id    	TEXT NOT NULL,
    hnm_id        	TEXT NOT NULL,
    last_kill     	TEXT NOT NULL,
    days_since_hq 	INTEGER NOT NULL,
	is_notified   	INTEGER NOT NULL DEFAULT 0,
	mod				TEXT NOT NULL,
    created_at    	TEXT NOT NULL,
    updated_at    	TEXT NOT NULL,
    UNIQUE (guild_id, channel_id, hnm_id)
);

CREATE TABLE IF NOT EXISTS hnm_timer_boards (
    guild_id   TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    PRIMARY KEY (guild_id, channel_id)
);

CREATE TABLE IF NOT EXISTS hnm_camp_channels (
    id              		TEXT PRIMARY KEY,
    guild_id        		TEXT NOT NULL,
    channel_id      		TEXT NOT NULL,
    hnm_id          		TEXT NOT NULL,
    last_kill       		TEXT NOT NULL,
    days_since_hq   		INTEGER NOT NULL,
    seq             		INTEGER NOT NULL,
	mod						TEXT NOT NULL,
    is_closed       		INTEGER NOT NULL DEFAULT 0,
    is_enraged      		INTEGER NOT NULL DEFAULT 0,
	is_spawned				INTEGER NOT NULL DEFAULT 0,
	is_hq				INTEGER NOT NULL DEFAULT 0,
    last_window_idx 		INTEGER NOT NULL DEFAULT 0,
	last_warned_window_idx 	INTEGER NOT NULL DEFAULT 0,
	last_cutoff_window_idx 	INTEGER NOT NULL DEFAULT 0,
    created_at      		TEXT NOT NULL,
    updated_at      		TEXT NOT NULL,
	move_scheduled  		INTEGER NOT NULL DEFAULT 0,
    UNIQUE (guild_id, hnm_id, last_kill, days_since_hq)
);

CREATE TABLE IF NOT EXISTS hnm_spawns (
	id				TEXT PRIMARY KEY,
	guild_id		TEXT NOT NULL,
	hnm_id			TEXT NOT NULL,
	hnm_name		TEXT NOT NULL,
	linkshell		TEXT NOT NULL,
	is_hq			INTEGER NOT NULL,
	window_idx		INTEGER NOT NULL,
	spawned_at 		TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS linkshells (
	id						TEXT PRIMARY KEY,
	guild_id				TEXT NOT NULL,
	linkshell_name			TEXT NOT NULL,
	fafnir_claims			INTEGER NOT NULL DEFAULT 0,
	adamantoise_claims		INTEGER NOT NULL DEFAULT 0,
	behemoth_claims			INTEGER NOT NULL DEFAULT 0,
	simurgh_claims			INTEGER NOT NULL DEFAULT 0,
	shikigami_weapon_claims	INTEGER NOT NULL DEFAULT 0,
	king_arthro_claims		INTEGER NOT NULL DEFAULT 0,
	king_vinegarroon_claims	INTEGER NOT NULL DEFAULT 0,
	bloodsucker_claims		INTEGER NOT NULL DEFAULT 0,
	tiamat_claims			INTEGER NOT NULL DEFAULT 0,
	vrtra_claims			INTEGER NOT NULL DEFAULT 0,
	jormungand_claims		INTEGER NOT NULL DEFAULT 0,
	nidhogg_claims			INTEGER NOT NULL DEFAULT 0,
	aspidochelone_claims	INTEGER NOT NULL DEFAULT 0,
	king_behemoth_claims	INTEGER NOT NULL DEFAULT 0,
	khimaira_claims			INTEGER NOT NULL DEFAULT 0,
	cerberus_claims			INTEGER NOT NULL DEFAULT 0,
	hydra_claims			INTEGER NOT NULL DEFAULT 0,
	created_at				TEXT NOT NULL,
	updated_at				TEXT NOT NULL,
	archived_at				TEXT NOT NULL,
	UNIQUE (guild_id, linkshell_name)
);

CREATE TABLE IF NOT EXISTS linkshells_archive (
	id						TEXT PRIMARY KEY,
	guild_id				TEXT NOT NULL,
	linkshell_name			TEXT NOT NULL,
	fafnir_claims			INTEGER NOT NULL DEFAULT 0,
	adamantoise_claims		INTEGER NOT NULL DEFAULT 0,
	behemoth_claims			INTEGER NOT NULL DEFAULT 0,
	simurgh_claims			INTEGER NOT NULL DEFAULT 0,
	shikigami_weapon_claims	INTEGER NOT NULL DEFAULT 0,
	king_arthro_claims		INTEGER NOT NULL DEFAULT 0,
	king_vinegarroon_claims	INTEGER NOT NULL DEFAULT 0,
	bloodsucker_claims		INTEGER NOT NULL DEFAULT 0,
	tiamat_claims			INTEGER NOT NULL DEFAULT 0,
	vrtra_claims			INTEGER NOT NULL DEFAULT 0,
	jormungand_claims		INTEGER NOT NULL DEFAULT 0,
	nidhogg_claims			INTEGER NOT NULL DEFAULT 0,
	aspidochelone_claims	INTEGER NOT NULL DEFAULT 0,
	king_behemoth_claims	INTEGER NOT NULL DEFAULT 0,
	khimaira_claims			INTEGER NOT NULL DEFAULT 0,
	cerberus_claims			INTEGER NOT NULL DEFAULT 0,
	hydra_claims			INTEGER NOT NULL DEFAULT 0,
	created_at				TEXT NOT NULL,
	updated_at				TEXT NOT NULL,
	archived_at				TEXT NOT NULL,
	UNIQUE (guild_id, linkshell_name)
);

CREATE TABLE IF NOT EXISTS linkshell_leader_boards (
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

	if err := migrateColumns(db); err != nil {
		return nil, err
	}

	return &Store{DB: db}, nil
}

func migrateColumns(db *sql.DB) error {
	tableColumns := map[string][]string{
		"linkshells":         {"khimaira_claims", "cerberus_claims", "hydra_claims"},
		"linkshells_archive": {"khimaira_claims", "cerberus_claims", "hydra_claims"},
	}

	for table, columns := range tableColumns {
		existing, err := existingColumns(db, table)
		if err != nil {
			return err
		}

		for _, col := range columns {
			if existing[col] {
				continue
			}
			stmt := "ALTER TABLE " + table + " ADD COLUMN " + col + " INTEGER NOT NULL DEFAULT 0"
			if _, err := db.Exec(stmt); err != nil {
				return err
			}
		}
	}

	return nil
}

func existingColumns(db *sql.DB, table string) (map[string]bool, error) {
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue sql.NullString
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			return nil, err
		}
		cols[name] = true
	}
	return cols, rows.Err()
}
