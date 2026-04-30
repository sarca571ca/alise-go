package data

import (
	"database/sql"
	"time"
)

type HNMCampChannel struct {
	ID            string
	GuildID       string
	ChannelID     string
	HNMID         string
	LastKill      time.Time
	DaysSinceHQ   int
	Seq           int
	IsClosed      bool
	IsEnraged     bool
	LastWindowIdx int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	MoveScheduled bool
}

func (s *Store) UpsertHNMCampChannel(ch HNMCampChannel) (HNMCampChannel, error) {
	now := time.Now().UTC()

	if ch.ID == "" {
		ch.ID = generateID()
	}
	if ch.CreatedAt.IsZero() {
		ch.CreatedAt = now
	}
	ch.UpdatedAt = now

	isClosed := 0
	if ch.IsClosed {
		isClosed = 1
	}

	isEnraged := 0
	if ch.IsEnraged {
		isEnraged = 1
	}

	moveScheduled := 0
	if ch.MoveScheduled {
		moveScheduled = 1
	}

	const q = `
	INSERT INTO hnm_camp_channels (
		id, guild_id, channel_id, hnm_id, last_kill, days_since_hq,
		seq, is_closed, is_enraged, last_window_idx, move_scheduled,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(guild_id, hnm_id, last_kill, days_since_hq) DO UPDATE SET
		last_kill       = excluded.last_kill,
		days_since_hq   = excluded.days_since_hq,
		seq             = excluded.seq,
		is_closed       = excluded.is_closed,
		is_enraged      = excluded.is_enraged,
		last_window_idx = excluded.last_window_idx,
		move_scheduled  = excluded.move_scheduled,
		updated_at      = excluded.updated_at
	`

	_, err := s.DB.Exec(q, ch.ID, ch.GuildID, ch.ChannelID, ch.HNMID, toStrTime(ch.LastKill),
		ch.DaysSinceHQ, ch.Seq, isClosed, isEnraged, ch.LastWindowIdx, moveScheduled,
		toStrTime(ch.CreatedAt), toStrTime(ch.UpdatedAt),
	)

	return ch, err
}

func (s *Store) GetHNMCampChannelByChannelID(guildID, channelID string) (HNMCampChannel, bool, error) {
	const q = `
	SELECT id, guild_id, channel_id, hnm_id, last_kill, days_since_hq,
		   seq, is_closed, is_enraged, last_window_idx, move_scheduled, created_at, updated_at 
	FROM hnm_camp_channels
	WHERE guild_id = ? AND channel_id = ?
	`

	var (
		ch                 HNMCampChannel
		lastKillStr        string
		isClosedInt        int
		isEnragedInt       int
		createdAtStr       string
		updatedAtStr       string
		isMoveScheduledInt int
	)

	err := s.DB.QueryRow(q, guildID, channelID).Scan(
		&ch.ID, &ch.GuildID, &ch.ChannelID, &ch.HNMID, &lastKillStr, &ch.DaysSinceHQ,
		&ch.Seq, &isClosedInt, &isEnragedInt, &ch.LastWindowIdx, &isMoveScheduledInt,
		&createdAtStr, &updatedAtStr,
	)

	if err == sql.ErrNoRows {
		return HNMCampChannel{}, false, nil
	}
	if err != nil {
		return HNMCampChannel{}, false, err
	}

	var parseErr error
	if ch.LastKill, parseErr = fromStrTime(lastKillStr); parseErr != nil {
		return HNMCampChannel{}, false, parseErr
	}
	if ch.CreatedAt, parseErr = fromStrTime(createdAtStr); parseErr != nil {
		return HNMCampChannel{}, false, parseErr
	}
	if ch.UpdatedAt, parseErr = fromStrTime(updatedAtStr); parseErr != nil {
		return HNMCampChannel{}, false, parseErr
	}
	ch.IsClosed = isClosedInt != 0
	ch.IsEnraged = isEnragedInt != 0
	ch.MoveScheduled = isMoveScheduledInt != 0

	return ch, true, nil
}

func (s *Store) ListHNMCampChannels(guildID string) ([]HNMCampChannel, error) {
	const q = `
	SELECT id, guild_id, channel_id, hnm_id, last_kill, days_since_hq,
		   seq, is_closed, is_enraged, last_window_idx, move_scheduled, created_at, updated_at
	FROM hnm_camp_channels
	WHERE guild_id = ?
	`

	rows, err := s.DB.Query(q, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []HNMCampChannel
	for rows.Next() {
		var (
			ch                 HNMCampChannel
			lastKillStr        string
			isClosedInt        int
			isEnragedInt       int
			createdAtStr       string
			updatedAtStr       string
			isMoveScheduledInt int
		)
		if err := rows.Scan(
			&ch.ID, &ch.GuildID, &ch.ChannelID, &ch.HNMID, &lastKillStr, &ch.DaysSinceHQ,
			&ch.Seq, &isClosedInt, &isEnragedInt, &ch.LastWindowIdx, &isMoveScheduledInt,
			&createdAtStr, &updatedAtStr,
		); err != nil {
			return nil, err
		}

		if ch.LastKill, err = fromStrTime(lastKillStr); err != nil {
			return nil, err
		}
		if ch.CreatedAt, err = fromStrTime(createdAtStr); err != nil {
			return nil, err
		}
		if ch.UpdatedAt, err = fromStrTime(updatedAtStr); err != nil {
			return nil, err
		}
		ch.IsClosed = isClosedInt != 0
		ch.IsEnraged = isEnragedInt != 0
		ch.MoveScheduled = isMoveScheduledInt != 0
		out = append(out, ch)
	}

	return out, rows.Err()
}

func (s *Store) ListHNMCampChannelsForDay(guildID, hnmID string, day time.Time) ([]HNMCampChannel, error) {
	// Normalize to date boundary in UTC (or your chosen TZ)
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		loc = time.UTC
	}

	day = day.In(loc)
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	const q = `
		SELECT id, guild_id, channel_id, hnm_id,
		       last_kill, days_since_hq, seq,
		       is_closed, is_enraged, last_window_idx,
		       move_scheduled, created_at, updated_at
		FROM hnm_camp_channels
		WHERE guild_id = ?
		  AND hnm_id = ?
		  AND last_kill >= ?
		  AND last_kill < ?
	`

	rows, err := s.DB.Query(q, guildID, hnmID, toStrTime(start), toStrTime(end))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []HNMCampChannel
	for rows.Next() {
		var (
			ch                 HNMCampChannel
			lastKillStr        string
			isClosedInt        int
			isEnragedInt       int
			createdAtStr       string
			updatedAtStr       string
			isMoveScheduledInt int
		)
		if err := rows.Scan(
			&ch.ID, &ch.GuildID, &ch.ChannelID, &ch.HNMID,
			&lastKillStr, &ch.DaysSinceHQ, &ch.Seq,
			&isClosedInt, &isEnragedInt, &ch.LastWindowIdx,
			&isMoveScheduledInt, &createdAtStr, &updatedAtStr,
		); err != nil {
			return nil, err
		}

		if ch.LastKill, err = fromStrTime(lastKillStr); err != nil {
			return nil, err
		}
		if ch.CreatedAt, err = fromStrTime(createdAtStr); err != nil {
			return nil, err
		}
		if ch.UpdatedAt, err = fromStrTime(updatedAtStr); err != nil {
			return nil, err
		}
		ch.IsClosed = isClosedInt != 0
		ch.IsEnraged = isEnragedInt != 0
		ch.MoveScheduled = isMoveScheduledInt != 0

		out = append(out, ch)
	}
	return out, rows.Err()
}

func (s *Store) GetHNMCampChannelByCamp(
	guildID, hnmID string,
	lastKill time.Time,
	daysSinceHQ int,
) (HNMCampChannel, bool, error) {
	const q = `
	SELECT id, guild_id, channel_id, hnm_id, last_kill, days_since_hq,
		   seq, is_closed, is_enraged, last_window_idx, move_scheduled, created_at, updated_at
	FROM hnm_camp_channels
	WHERE guild_id = ? AND hnm_id = ? AND last_kill = ? AND days_since_hq = ?
	`

	var (
		ch                 HNMCampChannel
		lastKillStr        string
		isClosedInt        int
		isEnragedInt       int
		createdAtStr       string
		updatedAtStr       string
		isMoveScheduledInt int
	)

	err := s.DB.QueryRow(
		q,
		guildID,
		hnmID,
		toStrTime(lastKill),
		daysSinceHQ,
	).Scan(
		&ch.ID, &ch.GuildID, &ch.ChannelID, &ch.HNMID, &lastKillStr, &ch.DaysSinceHQ,
		&ch.Seq, &isClosedInt, &isEnragedInt, &ch.LastWindowIdx, &isMoveScheduledInt,
		&createdAtStr, &updatedAtStr,
	)
	if err == sql.ErrNoRows {
		return HNMCampChannel{}, false, nil
	}
	if err != nil {
		return HNMCampChannel{}, false, err
	}

	var parseErr error
	if ch.LastKill, parseErr = fromStrTime(lastKillStr); parseErr != nil {
		return HNMCampChannel{}, false, parseErr
	}
	if ch.CreatedAt, parseErr = fromStrTime(createdAtStr); parseErr != nil {
		return HNMCampChannel{}, false, parseErr
	}
	if ch.UpdatedAt, parseErr = fromStrTime(updatedAtStr); parseErr != nil {
		return HNMCampChannel{}, false, parseErr
	}
	ch.IsClosed = isClosedInt != 0
	ch.IsEnraged = isEnragedInt != 0
	ch.MoveScheduled = isMoveScheduledInt != 0

	return ch, true, nil
}

func (s *Store) DeleteHNMCampChannel(id string) error {
	const q = `DELETE FROM hnm_camp_channels WHERE id = ?`
	_, err := s.DB.Exec(q, id)
	return err
}
