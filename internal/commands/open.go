package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type OpenCommand struct{}

func (OpenCommand) Name() string        { return "open" }
func (OpenCommand) Description() string { return "Opens camp channel after closed." }
func (OpenCommand) SlashDef() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Name:        "open",
		Description: "Opens camp channel after closed.",
	}
}
func (cmd OpenCommand) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Println("Open")
}
