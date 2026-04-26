package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type PopCommand struct{}

func (PopCommand) Name() string        { return "open" }
func (PopCommand) Description() string { return "Opens camp channel after closed." }
func (PopCommand) SlashDef() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Name:        "open",
		Description: "Opens camp channel after closed.",
	}
}
func (cmd PopCommand) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Println("Pop")
}
