package services

import (
	"alise-go/internal/config"
	"alise-go/internal/data"

	"github.com/bwmarrin/discordgo"
)

// TODO: Service will be for making and maintaining the Linkshell list for AutoComplete
// NOTE: This is on hold. Future QoL change and its not required at this time.

type LinkshellService struct {
	store *data.Store
	cfg   config.Config
	dg    *discordgo.Session
}

func NewLinkshellService(store *data.Store, cfg config.Config, dg *discordgo.Session) *HNMService {
	return &HNMService{
		store: store,
		cfg:   cfg,
		dg:    dg,
	}
}
