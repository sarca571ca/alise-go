package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type EnrageCommand struct{}

func (EnrageCommand) Name() string        { return "open" }
func (EnrageCommand) Description() string { return "Opens camp channel after closed." }
func (EnrageCommand) SlashDef() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Name:        "open",
		Description: "Opens camp channel after closed.",
	}
}
func (cmd EnrageCommand) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Println("Enrage")
}
