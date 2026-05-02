package services

import (
	"alise-go/internal/config"
	"alise-go/internal/data"
	"alise-go/internal/formatting"
	"alise-go/internal/models"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

type CampService struct {
	store *data.Store
	cfg   config.Config
	dg    *discordgo.Session
}

func NewCampService(store *data.Store, cfg config.Config, dg *discordgo.Session) *CampService {
	return &CampService{
		store: store,
		cfg:   cfg,
		dg:    dg,
	}
}

func (s *CampService) Pop(guildID, lsName string, ch *discordgo.Channel) (data.HNMCampChannel, error) {
	if err := s.validateCampChannel(*ch); err != nil {
		return data.HNMCampChannel{}, err
	}

	if !s.isHNMCategory(*ch) {
		return data.HNMCampChannel{}, fmt.Errorf("this command can only be used in HNM camp channels")
	}

	camp, ok, err := s.store.GetHNMCampChannelByChannelID(guildID, ch.ID)
	if err != nil {
		return data.HNMCampChannel{}, fmt.Errorf("failed to load camp")
	}
	if !ok {
		return data.HNMCampChannel{}, fmt.Errorf("no camp record found for this channel")
	}

	if camp.IsSpawned {
		return data.HNMCampChannel{}, fmt.Errorf("pop has already been recorded for this camp")
	}

	hnm, ok := models.GetHNM(camp.HNMID)
	if !ok {
		return data.HNMCampChannel{}, fmt.Errorf("failed to resolve HNM for this camp")
	}

	rec, ok, err := s.store.GetLinkshellRecord(guildID, lsName)
	if err != nil {
		return data.HNMCampChannel{}, fmt.Errorf("failed to load linkshell")
	}
	if !ok {
		return data.HNMCampChannel{}, fmt.Errorf("linkshell %q does not exist, use /linkshell list to see available options", lsName)
	}

	rec, err = s.incrementLinkshellClaim(rec, hnm.ID)
	if err != nil {
		return data.HNMCampChannel{}, err
	}

	if _, err := s.store.UpsertLinkshellRecord(rec); err != nil {
		return data.HNMCampChannel{}, fmt.Errorf("failed to update linkshell claim count")
	}

	camp.IsSpawned = true
	camp.MoveScheduled = true

	if _, err := s.store.UpsertHNMCampChannel(camp); err != nil {
		return data.HNMCampChannel{}, fmt.Errorf("failed to update camp state")
	}

	return camp, nil
}

func (s *CampService) validateCampChannel(ch discordgo.Channel) error {
	_, ok, err := s.store.GetHNMCampChannelByChannelID(s.cfg.GuildID, ch.ID)
	if !ok {
		return fmt.Errorf("Camp not found in hnm_camp_channels table.")
	}
	if err != nil {
		return err
	}

	return nil
}

func (s *CampService) isHNMCategory(ch discordgo.Channel) bool {
	return ch.ParentID == s.cfg.Categories.HNMCategoryID
}

func (s *CampService) incrementLinkshellClaim(
	rec data.LinkshellRecord,
	hnmID string,
) (data.LinkshellRecord, error) {
	switch hnmID {
	case "fafnir":
		rec.FafnirClaims++
	case "adamantoise":
		rec.AdamantoiseClaims++
	case "behemoth":
		rec.BehemothClaims++
	case "tiamat":
		rec.TiamatClaims++
	case "jorm":
		rec.JormungandClaims++
	case "vrtra":
		rec.VrtraClaims++
	case "simurgh":
		rec.SimurghClaims++
	case "ka":
		rec.KingArthroClaims++
	case "bloodsucker":
		rec.BloodsuckerClaims++
	case "shiki":
		rec.ShikigamiWeaponClaims++
	case "kv":
		rec.KingVinegarroonClaims++
	default:
		return rec, fmt.Errorf("unsupported HNM id: %q", hnmID)
	}

	return rec, nil
}

func (s *CampService) EnrageWindow(channelID string, window int, delay time.Duration, camp data.HNMCampChannel) {
	time.Sleep(delay)

	camp.IsEnraged = false

	if _, err := s.dg.ChannelMessageSend(
		channelID,
		fmt.Sprintf("%v\nMoving channel to awaiting-processing in 5 minutes.", formatting.FormatWindowHeading(fmt.Sprintf("Enrage %v Closed", window))),
	); err != nil {
		return
	}

	if _, err := s.store.UpsertHNMCampChannel(camp); err != nil {
		return
	}
}
