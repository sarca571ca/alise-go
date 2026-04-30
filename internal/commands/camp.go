package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

type PopHandler func(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	ls string,
)

type OpenHandler func(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
)

type CloseHandler func(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
)

type EnrageHandler func(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
)

type StableHandler func(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
)

type CTAHandler func(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
)
type CampCommand struct {
	Pop    PopHandler
	Open   OpenHandler
	Close  CloseHandler
	Enrage EnrageHandler
	Stable StableHandler
	CTA    CTAHandler
}

func (CampCommand) Name() string        { return "camp" }
func (CampCommand) Description() string { return "Changes the status of the camp." }
func (CampCommand) SlashDef() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "camp",
		Description: "Changes the status of the camp.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "pop",
				Description: "Ends the camp and saves the Linkshell that claimed.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:         "name",
						Description:  "Linkshell name that claimed.",
						Type:         discordgo.ApplicationCommandOptionString,
						Required:     true,
						Autocomplete: true, // NOTE: Need to autocomplete from the linkshells table in the db
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "open",
				Description: "Opens the camp after closing.",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "close",
				Description: "Closes the camp.",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "enrage",
				Description: "Re-Activates the camp if the HNM enrages.",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "stable",
				Description: "Ends the camp after enrage.",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "cta",
				Description: "Creates a Call to Arms version of the current camp.",
			},
		},
	}
}

func (cmd CampCommand) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	if len(data.Options) == 0 {
		return
	}

	sub := data.Options[0]
	switch sub.Name {
	case "pop":
		cmd.handlePop(s, i, sub)
	case "open":
		cmd.handleOpen(s, i)
	case "close":
		cmd.handleClose(s, i)
	case "enrage":
		cmd.handleEnrage(s, i)
	case "stable":
		cmd.handleStable(s, i)
	case "cta":
		cmd.handleCTA(s, i)
	}
}

// TODO: AutoComplete from linkshells tabl"e
func (cmd CampCommand) HandleAutoCompleteCommand(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	focused *discordgo.ApplicationCommandInteractionDataOption,
) {
	partial := strings.ToLower(focused.StringValue())

	println(partial)
}

func (cmd CampCommand) handlePop(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	sub *discordgo.ApplicationCommandInteractionDataOption,
) {
	cmd.Pop(s, i, getNameOption(sub))
}

func (cmd CampCommand) handleOpen(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) {
	cmd.Open(s, i)
}

func (cmd CampCommand) handleClose(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) {
	cmd.Close(s, i)
}

func (cmd CampCommand) handleEnrage(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) {
	cmd.Enrage(s, i)
}

func (cmd CampCommand) handleStable(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) {
	cmd.Stable(s, i)
}

func (cmd CampCommand) handleCTA(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) {
	cmd.CTA(s, i)
}
