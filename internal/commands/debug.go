package commands

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type DebugOnParsedHandler func(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	window string,
)

type DebugCommand struct {
	OnParsed DebugOnParsedHandler
}

func (DebugCommand) Name() string        { return "debug" }
func (DebugCommand) Description() string { return "Debugging only" }
func (DebugCommand) SlashDef() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "debug",
		Description: "Debugging only",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:         "window",
				Description:  "window",
				Type:         discordgo.ApplicationCommandOptionString,
				Required:     true,
				Autocomplete: false,
			},
		},
	}
}

func (cmd DebugCommand) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	windowOpts := data.GetOption("window")
	window := strings.TrimSpace(strings.ToLower(windowOpts.Value.(string)))
	log.Println(window)
	cmd.OnParsed(s, i, window)
}
