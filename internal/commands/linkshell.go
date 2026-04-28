package commands

import (
	"github.com/bwmarrin/discordgo"
)

type LinkshellListHandler func(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
)

type LinkshellAddHandler func(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	ls string,
)

type LinkshellRemoveHandler func(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	ls string,
)

type LinkshellUpdateHandler func(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	ls string,
)

type LinkshellCommand struct {
	List   LinkshellListHandler
	Add    LinkshellAddHandler
	Remove LinkshellRemoveHandler
	Update LinkshellUpdateHandler
}

func (LinkshellCommand) Name() string        { return "linkshell" }
func (LinkshellCommand) Description() string { return "Manages Linkshell DB records." }
func (LinkshellCommand) SlashDef() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "linkshell",
		Description: "Manages Linkshell DB records.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "list",
				Description: "Lists linkshells that are saved to record.",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Adds a linkshell to record.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:         "name",
						Description:  "Linkshell name (eg. WhereDragon, Hideout)",
						Type:         discordgo.ApplicationCommandOptionString,
						Required:     true,
						Autocomplete: false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Removes a linkshell from records.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:         "name",
						Description:  "Linkshell name (eg. WhereDragon, Hideout)",
						Type:         discordgo.ApplicationCommandOptionString,
						Required:     true,
						Autocomplete: false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "restore",
				Description: "Restores a linkshell from archive.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:         "name",
						Description:  "Linkshell name (eg. WhereDragon, Hideout)",
						Type:         discordgo.ApplicationCommandOptionString,
						Required:     true,
						Autocomplete: false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "update",
				Description: "Updates a linkshells claim record.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:         "name",
						Description:  "Linkshell name (eg. WhereDragon, Hideout)",
						Type:         discordgo.ApplicationCommandOptionString,
						Required:     true,
						Autocomplete: false,
					},
				},
			},
		},
	}
}

func (cmd LinkshellCommand) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	if len(data.Options) == 0 {
		return
	}

	sub := data.Options[0]
	switch sub.Name {
	case "list":
		cmd.handleList(s, i)
	case "add":
		cmd.handleAdd(s, i, sub)
	case "remove":
		cmd.handleRemove(s, i, sub)
	case "restore":
		cmd.handleRestore(s, i, sub)
	case "update":
		cmd.handleUpdate(s, i, sub)
	}

}

func (cmd LinkshellCommand) handleList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmd.List(s, i)
}

func (cmd LinkshellCommand) handleAdd(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	sub *discordgo.ApplicationCommandInteractionDataOption,
) {
	cmd.Add(s, i, getNameOption(sub))
}

func (cmd LinkshellCommand) handleRemove(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	sub *discordgo.ApplicationCommandInteractionDataOption,
) {
	cmd.Remove(s, i, getNameOption(sub))
}

func (cmd LinkshellCommand) handleRestore(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	sub *discordgo.ApplicationCommandInteractionDataOption,
) {
	// Restore
	// TODO: Restores the Linkshell record from the archive and returns an ephemeral.
}

func (cmd LinkshellCommand) handleUpdate(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	sub *discordgo.ApplicationCommandInteractionDataOption,
) {
	// update
	// TODO: shows the user a modal of the linkshell record. each field is prefilled
	// with their current values and maybe show the current value in the label to allow
	// for mistakes. allow for updating values as required. log the user
	// so we know who is modifying records
	cmd.Update(s, i, getNameOption(sub))
}

func getNameOption(sub *discordgo.ApplicationCommandInteractionDataOption) string {
	for _, opt := range sub.Options {
		if opt.Name == "name" {
			return opt.StringValue()
		}
	}
	return ""
}
