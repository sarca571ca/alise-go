package services

import (
	"alise-go/internal/config"
	"alise-go/internal/data"
	"fmt"

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

func (s *CampService) validateCampChannel(ch discordgo.Channel) error {
	//TODO: Searched the hnm_camp_channels table in the db by channel id to see if its a
	// valid camp
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
