package data

import (
	"alise-go/internal/models"
	"database/sql"
	"time"
)

type LinkshellRecord struct {
	ID                    string
	GuildID               string
	LinkshellName         string
	FafnirClaims          int
	AdamantoiseClaims     int
	BehemothClaims        int
	SimurghClaims         int
	ShikigamiWeaponClaims int
	KingArthroClaims      int
	KingVinegarroonClaims int
	BloodsuckerClaims     int
	TiamatClaims          int
	VrtraClaims           int
	JormungandClaims      int
	NidhoggClaims         int
	AspidocheloneClaims   int
	KingBehemothClaims    int
	CreatedAt             time.Time
	UpdatedAt             time.Time
	ArchivedAt            time.Time
}

func (s *Store) NewRecordFromLinkshellEntry(guildID string, cmdInput string) LinkshellRecord {
	return LinkshellRecord{
		GuildID:       guildID,
		LinkshellName: cmdInput,
	}
}

func (s *Store) NewLinkshellListFromRecords(linkshells []LinkshellRecord) []models.Linkshell {
	var linkshellList []models.Linkshell

	for _, ls := range linkshells {
		linkshellList = append(linkshellList, models.Linkshell{
			LinkshellName:         ls.LinkshellName,
			FafnirClaims:          ls.FafnirClaims,
			AdamantoiseClaims:     ls.AdamantoiseClaims,
			BehemothClaims:        ls.BehemothClaims,
			SimurghClaims:         ls.SimurghClaims,
			ShikigamiWeaponClaims: ls.ShikigamiWeaponClaims,
			KingArthroClaims:      ls.KingArthroClaims,
			KingVinegarroonClaims: ls.KingVinegarroonClaims,
			BloodsuckerClaims:     ls.BloodsuckerClaims,
			TiamatClaims:          ls.TiamatClaims,
			VrtraClaims:           ls.VrtraClaims,
			JormungandClaims:      ls.JormungandClaims,
			NidhoggClaims:         ls.NidhoggClaims,
			AspidocheloneClaims:   ls.AspidocheloneClaims,
			KingBehemothClaims:    ls.KingBehemothClaims,
		})
	}
	return linkshellList
}

func (s *Store) UpsertLinkshellRecord(ls LinkshellRecord) (LinkshellRecord, error) {
	now := time.Now().UTC()

	if ls.ID == "" {
		ls.ID = generateID()
	}
	if ls.CreatedAt.IsZero() {
		ls.CreatedAt = now
	}
	archivedAt := ""
	if !ls.ArchivedAt.IsZero() {
		archivedAt = toStrTime(ls.ArchivedAt)
	}
	ls.UpdatedAt = now

	const q = `
	INSERT INTO linkshells (
		id, guild_id, linkshell_name, fafnir_claims, adamantoise_claims,
		behemoth_claims, simurgh_claims, shikigami_weapon_claims,
		king_arthro_claims, king_vinegarroon_claims, bloodsucker_claims,
		tiamat_claims, vrtra_claims, jormungand_claims, nidhogg_claims,
		aspidochelone_claims, king_behemoth_claims, created_at, updated_at,
		archived_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(guild_id, linkshell_name) DO UPDATE SET
		fafnir_claims			= excluded.fafnir_claims,
		adamantoise_claims		= excluded.adamantoise_claims,
		behemoth_claims 		= excluded.behemoth_claims,
		simurgh_claims 			= excluded.simurgh_claims,
		shikigami_weapon_claims = excluded.shikigami_weapon_claims,
		king_arthro_claims		= excluded.king_arthro_claims,
		king_vinegarroon_claims = excluded.king_vinegarroon_claims,
		bloodsucker_claims 		= excluded.bloodsucker_claims,
		tiamat_claims			= excluded.tiamat_claims,
		vrtra_claims			= excluded.vrtra_claims,
		jormungand_claims		= excluded.jormungand_claims,
		nidhogg_claims			= excluded.nidhogg_claims,
		aspidochelone_claims	= excluded.aspidochelone_claims,
		king_behemoth_claims	= excluded.king_behemoth_claims,
		updated_at      		= excluded.updated_at
	`

	_, err := s.DB.Exec(
		q, ls.ID, ls.GuildID, ls.LinkshellName, ls.FafnirClaims, ls.AdamantoiseClaims,
		ls.BehemothClaims, ls.SimurghClaims, ls.ShikigamiWeaponClaims, ls.KingArthroClaims,
		ls.KingVinegarroonClaims, ls.BloodsuckerClaims, ls.TiamatClaims, ls.VrtraClaims,
		ls.JormungandClaims, ls.NidhoggClaims, ls.AspidocheloneClaims, ls.KingBehemothClaims,
		toStrTime(ls.CreatedAt), toStrTime(ls.UpdatedAt), archivedAt,
	)

	return ls, err
}

// ListLinkshellRecord
func (s *Store) ListLinkshellRecords(guildID string) ([]LinkshellRecord, error) {
	const q = `
	SELECT id, guild_id, linkshell_name, fafnir_claims, aspidochelone_claims,
	behemoth_claims, simurgh_claims, shikigami_weapon_claims, king_arthro_claims,
	king_vinegarroon_claims, bloodsucker_claims, tiamat_claims, vrtra_claims,
	jormungand_claims, nidhogg_claims, aspidochelone_claims, king_behemoth_claims,
	created_at, updated_at, archived_at
	FROM linkshells
	WHERE guild_id = ?
	`
	rows, err := s.DB.Query(q, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LinkshellRecord
	for rows.Next() {
		var (
			ls            LinkshellRecord
			createdAtStr  string
			updatedAtStr  string
			archivedAtStr string
		)
		if err := rows.Scan(
			&ls.ID, &ls.GuildID, &ls.LinkshellName, &ls.FafnirClaims, &ls.AdamantoiseClaims,
			&ls.BehemothClaims, &ls.SimurghClaims, &ls.ShikigamiWeaponClaims, &ls.KingArthroClaims,
			&ls.KingVinegarroonClaims, &ls.BloodsuckerClaims, &ls.TiamatClaims, &ls.VrtraClaims,
			&ls.JormungandClaims, &ls.NidhoggClaims, &ls.AspidocheloneClaims, &ls.KingBehemothClaims,
			&createdAtStr, &updatedAtStr, &archivedAtStr,
		); err != nil {
			return nil, err
		}
		if ls.CreatedAt, err = fromStrTime(createdAtStr); err != nil {
			return nil, err
		}
		if ls.UpdatedAt, err = fromStrTime(updatedAtStr); err != nil {
			return nil, err
		}
		if archivedAtStr != "" {
			if ls.ArchivedAt, err = fromStrTime(archivedAtStr); err != nil {
				return nil, err
			}
		}

		out = append(out, ls)
	}

	return out, rows.Err()
}

// GetLinkshellRecord
func (s *Store) GetLinkshellRecord(guildID, linkshellName string) (LinkshellRecord, bool, error) {
	const q = `
	SELECT id, guild_id, linkshell_name, fafnir_claims, aspidochelone_claims,
	behemoth_claims, simurgh_claims, shikigami_weapon_claims, king_arthro_claims,
	king_vinegarroon_claims, bloodsucker_claims, tiamat_claims, vrtra_claims,
	jormungand_claims, nidhogg_claims, aspidochelone_claims, king_behemoth_claims,
	created_at, updated_at, archived_at
	FROM linkshells
	WHERE guild_id = ? AND linkshell_name = ?
	`

	var (
		ls            LinkshellRecord
		createdAtStr  string
		updatedAtStr  string
		archivedAtStr string
	)

	err := s.DB.QueryRow(q, guildID, linkshellName).Scan(
		&ls.ID, &ls.GuildID, &ls.LinkshellName, &ls.FafnirClaims, &ls.AdamantoiseClaims,
		&ls.BehemothClaims, &ls.SimurghClaims, &ls.ShikigamiWeaponClaims, &ls.KingArthroClaims,
		&ls.KingVinegarroonClaims, &ls.BloodsuckerClaims, &ls.TiamatClaims, &ls.VrtraClaims,
		&ls.JormungandClaims, &ls.NidhoggClaims, &ls.AspidocheloneClaims, &ls.KingBehemothClaims,
		&createdAtStr, &updatedAtStr, &archivedAtStr,
	)
	if err == sql.ErrNoRows {
		return LinkshellRecord{}, false, nil
	}
	if err != nil {
		return LinkshellRecord{}, false, err
	}

	var parseErr error
	if ls.CreatedAt, parseErr = fromStrTime(createdAtStr); parseErr != nil {
		return LinkshellRecord{}, false, parseErr
	}
	if ls.UpdatedAt, parseErr = fromStrTime(updatedAtStr); parseErr != nil {
		return LinkshellRecord{}, false, parseErr
	}
	if archivedAtStr != "" {
		if ls.ArchivedAt, err = fromStrTime(archivedAtStr); err != nil {
			return LinkshellRecord{}, false, parseErr
		}
	}

	return ls, true, err
}

// DeleteLinkshellRecord
func (s *Store) DeleteLinkshellRecord(guildID, linkshellName string) error {
	// Need to call ArchiveLinkshellRecord first to allow for mistakes
	ls, archiveErr := s.ArchiveLinkshellRecord(guildID, linkshellName)
	if archiveErr != nil {
		return archiveErr
	}
	const q = `DELETE FROM linkshells WHERE guild_id = ? AND linkshell_name = ?`
	_, err := s.DB.Exec(q, ls.GuildID, ls.LinkshellName)
	return err
}

// ArchiveLinkshellRecord - Will allow for undoing deletions that were accidental
func (s *Store) ArchiveLinkshellRecord(guildID, linkshellName string) (LinkshellRecord, error) {
	ls, exists, err := s.GetLinkshellRecord(guildID, linkshellName)
	if !exists {
		return LinkshellRecord{}, err
	}
	if err != nil {
		return LinkshellRecord{}, err
	}

	now := time.Now().UTC()

	if ls.ArchivedAt.IsZero() {
		ls.ArchivedAt = now
	}
	ls.UpdatedAt = now

	const q = `
	INSERT INTO linkshells_archive (
		id, guild_id, linkshell_name, fafnir_claims, adamantoise_claims,
		behemoth_claims, simurgh_claims, shikigami_weapon_claims,
		king_arthro_claims, king_vinegarroon_claims, bloodsucker_claims,
		tiamat_claims, vrtra_claims, jormungand_claims, nidhogg_claims,
		aspidochelone_claims, king_behemoth_claims, created_at, updated_at,
		archived_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(guild_id, linkshell_name) DO UPDATE SET
		fafnir_claims			= excluded.fafnir_claims,
		adamantoise_claims		= excluded.adamantoise_claims,
		behemoth_claims 		= excluded.behemoth_claims,
		simurgh_claims 			= excluded.simurgh_claims,
		shikigami_weapon_claims = excluded.shikigami_weapon_claims,
		king_arthro_claims		= excluded.king_arthro_claims,
		king_vinegarroon_claims = excluded.king_vinegarroon_claims,
		bloodsucker_claims 		= excluded.bloodsucker_claims,
		tiamat_claims			= excluded.tiamat_claims,
		vrtra_claims			= excluded.vrtra_claims,
		jormungand_claims		= excluded.jormungand_claims,
		nidhogg_claims			= excluded.nidhogg_claims,
		aspidochelone_claims	= excluded.aspidochelone_claims,
		king_behemoth_claims	= excluded.king_behemoth_claims,
		updated_at      		= excluded.updated_at
	`

	_, execErr := s.DB.Exec(
		q, ls.ID, ls.GuildID, ls.LinkshellName, ls.FafnirClaims, ls.AdamantoiseClaims,
		ls.BehemothClaims, ls.SimurghClaims, ls.ShikigamiWeaponClaims, ls.KingArthroClaims,
		ls.KingVinegarroonClaims, ls.BloodsuckerClaims, ls.TiamatClaims, ls.VrtraClaims,
		ls.JormungandClaims, ls.NidhoggClaims, ls.AspidocheloneClaims, ls.KingBehemothClaims,
		toStrTime(ls.CreatedAt), toStrTime(ls.UpdatedAt), toStrTime(ls.ArchivedAt),
	)

	return ls, execErr
}

// TODO: We have archive capabilities atm but in the future will need to add the ability to restore
// from the archived table. This is fluff for now. Will need to definetely add soem sort of robust
// logging to track users incase of malitiously removing records to the archive.
func (s *Store) RestoreLinkshellRecord(guildID, linkshellName string) (LinkshellRecord, error) {
	const q = `
		SELECT id, guild_id, linkshell_name,
		       fafnir_claims, adamantoise_claims, behemoth_claims,
		       simurgh_claims, shikigami_weapon_claims, king_arthro_claims,
		       king_vinegarroon_claims, bloodsucker_claims, tiamat_claims,
		       vrtra_claims, jormungand_claims, nidhogg_claims,
		       aspidochelone_claims, king_behemoth_claims,
		       created_at, updated_at, archived_at
		FROM linkshells_archive
		WHERE guild_id = ? AND linkshell_name = ?
		ORDER BY archived_at DESC
		LIMIT 1
	`
	var (
		ls            LinkshellRecord
		createdAtStr  string
		updatedAtStr  string
		archivedAtStr string
	)
	err := s.DB.QueryRow(q, guildID, linkshellName).Scan(
		&ls.ID, &ls.GuildID, &ls.LinkshellName,
		&ls.FafnirClaims, &ls.AdamantoiseClaims, &ls.BehemothClaims,
		&ls.SimurghClaims, &ls.ShikigamiWeaponClaims, &ls.KingArthroClaims,
		&ls.KingVinegarroonClaims, &ls.BloodsuckerClaims, &ls.TiamatClaims,
		&ls.VrtraClaims, &ls.JormungandClaims, &ls.NidhoggClaims,
		&ls.AspidocheloneClaims, &ls.KingBehemothClaims,
		&createdAtStr, &updatedAtStr, &archivedAtStr,
	)
	if err == sql.ErrNoRows {
		return LinkshellRecord{}, nil
	}
	if err != nil {
		return LinkshellRecord{}, err
	}

	if ls.CreatedAt, err = fromStrTime(createdAtStr); err != nil {
		return LinkshellRecord{}, err
	}
	if ls.UpdatedAt, err = fromStrTime(updatedAtStr); err != nil {
		return LinkshellRecord{}, err
	}
	if ls.ArchivedAt, err = fromStrTime(archivedAtStr); err != nil {
		return LinkshellRecord{}, err
	}

	ls.ID = "" // UpsertLinkshellRecord will generate a new ID
	ls.ArchivedAt = time.Time{}

	return s.UpsertLinkshellRecord(ls)
}

// TODO: Still need to call all this appropriately with the linkshell command using a
// switch statement for the different mods.
