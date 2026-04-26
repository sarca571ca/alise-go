package data

import "time"

type HNMSpawn struct {
	ID        string
	GuildID   string
	HNMID     string
	HNMName   string
	Linkshell string
	IsHQ      bool
	WindowIdx int
	SpawnedAt time.Time
}

func (s *Store) InsertHNMSpawn(sp HNMSpawn) (HNMSpawn, error) {
	if sp.ID == "" {
		sp.ID = generateID()
	}
	if sp.SpawnedAt.IsZero() {
		sp.SpawnedAt = time.Now().UTC()
	}

	isHQ := 0
	if sp.IsHQ {
		isHQ = 1
	}

	const q = `
		INSERT INTO hnm_spawns (
			id, guild_id, hnm_id, hnm_name, linkshell, is_hq,
			window_idx, spawned_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.DB.Exec(
		q, sp.ID, sp.GuildID, sp.HNMID, sp.HNMName, sp.Linkshell,
		isHQ, sp.WindowIdx, toStrTime(sp.SpawnedAt),
	)

	return sp, err
}
