package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type CloseCommand struct{}

func (CloseCommand) Name() string        { return "open" }
func (CloseCommand) Description() string { return "Opens camp channel after closed." }
func (CloseCommand) SlashDef() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Name:        "open",
		Description: "Opens camp channel after closed.",
	}
}
func (cmd CloseCommand) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Println("close")
}
