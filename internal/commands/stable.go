package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type StableCommand struct{}

func (StableCommand) Name() string        { return "open" }
func (StableCommand) Description() string { return "Opens camp channel after closed." }
func (StableCommand) SlashDef() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Name:        "open",
		Description: "Opens camp channel after closed.",
	}
}
func (cmd StableCommand) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Println("Stable")
}
