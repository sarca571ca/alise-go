package data

import (
	"alise-go/internal/models"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type HNMTimerRecord struct {
	ID          string
	GuildID     string
	ChannelID   string
	HNMID       string
	LastKill    time.Time
	DaysSinceHQ int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type HNMTimerBoard struct {
	GuildID   string
	ChannelID string
	MessageID string
}

type Store struct {
	DB *sql.DB
}

func NewTimerFromRecord(rec HNMTimerRecord, hnm models.HNM) models.HNMTimer {
	return models.HNMTimer{
		HNM:         hnm,
		LastKill:    rec.LastKill,
		DaysSinceHQ: rec.DaysSinceHQ,
	}
}

func toStrTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func fromStrTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, s)
}

func generateID() string {
	return uuid.NewString()
}

func NewRecordFromHNMTimer(guildID, channelID string, t models.HNMTimer) HNMTimerRecord {
	return HNMTimerRecord{
		GuildID:     guildID,
		ChannelID:   channelID,
		HNMID:       t.HNM.ID,
		LastKill:    t.LastKill,
		DaysSinceHQ: t.DaysSinceHQ,
	}
}

func (s *Store) UpsertHNMTimerRecord(rec HNMTimerRecord) (HNMTimerRecord, error) {
	now := time.Now().UTC()

	if rec.ID == "" {
		rec.ID = generateID()
	}
	if rec.CreatedAt.IsZero() {
		rec.CreatedAt = now
	}
	rec.UpdatedAt = now

	const q = `
		INSERT INTO hnm_timers (
			id, guild_id, channel_id, hnm_id, last_kill, days_since_hq, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(guild_id, channel_id, hnm_id) DO UPDATE SET
			last_kill    = excluded.last_kill,
			days_since_hq = excluded.days_since_hq,
			updated_at   = excluded.updated_at
	`

	_, err := s.DB.Exec(q,
		rec.ID, rec.GuildID, rec.ChannelID, rec.HNMID,
		toStrTime(rec.LastKill), rec.DaysSinceHQ,
		toStrTime(rec.CreatedAt), toStrTime(rec.UpdatedAt),
	)
	return rec, err
}

func (s *Store) GetHNMTimerRecord(guildID, channelID, hnmID string) (HNMTimerRecord, bool, error) {
	const q = `
	SELECT id, guild_id, channel_id, hnm_id, last_kill, days_since_hq, created_at, updated_at
	FROM hnm_timers
	WHERE guild_id = ? AND channel_id = ? AND hnm_id = ?
	`

	var (
		rec          HNMTimerRecord
		lastKillStr  string
		createdAtStr string
		updatedAtStr string
	)

	err := s.DB.QueryRow(q, guildID, channelID, hnmID).Scan(
		&rec.ID, &rec.GuildID, &rec.ChannelID, &rec.HNMID,
		&lastKillStr, &rec.DaysSinceHQ, &createdAtStr, &updatedAtStr,
	)

	if err == sql.ErrNoRows {
		return HNMTimerRecord{}, false, nil
	}

	if err != nil {
		return HNMTimerRecord{}, false, err
	}

	if rec.LastKill, err = fromStrTime(lastKillStr); err != nil {
		return HNMTimerRecord{}, false, err
	}

	if rec.CreatedAt, err = fromStrTime(createdAtStr); err != nil {
		return HNMTimerRecord{}, false, err
	}

	if rec.UpdatedAt, err = fromStrTime(updatedAtStr); err != nil {
		return HNMTimerRecord{}, false, err
	}

	return rec, true, nil
}

func (s *Store) ListHNMTimerRecords(guildID, channelID string) ([]HNMTimerRecord, error) {
	const q = `
		SELECT id, guild_id, channel_id, hnm_id, last_kill, days_since_hq, created_at, updated_at
		FROM hnm_timers
		WHERE guild_id = ? AND channel_id = ?
		ORDER BY hnm_id
	`

	rows, err := s.DB.Query(q, guildID, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []HNMTimerRecord
	for rows.Next() {
		var (
			rec          HNMTimerRecord
			lastKillStr  string
			createdAtStr string
			updatedAtStr string
		)
		if err := rows.Scan(
			&rec.ID, &rec.GuildID, &rec.ChannelID, &rec.HNMID,
			&lastKillStr, &rec.DaysSinceHQ, &createdAtStr, &updatedAtStr,
		); err != nil {
			return nil, err
		}
		if rec.LastKill, err = fromStrTime(lastKillStr); err != nil {
			return nil, err
		}
		if rec.CreatedAt, err = fromStrTime(createdAtStr); err != nil {
			return nil, err
		}
		if rec.UpdatedAt, err = fromStrTime(updatedAtStr); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (s *Store) GetHNMTimerBoard(guildID, channelID string) (HNMTimerBoard, bool, error)
func (s *Store) UpsertHNMTimerBoard(board HNMTimerBoard) error
