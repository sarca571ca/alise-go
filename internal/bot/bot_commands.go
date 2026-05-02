package bot

import (
	"alise-go/internal/commands"
	"alise-go/internal/config"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) buildCommands(cfg config.Config) []commands.Command {
	return []commands.Command{
		b.withCommandLogging(b.buildHnmCommand(cfg)),
		b.withCommandLogging(b.buildLinkshellCommand(cfg)),
		b.withCommandLogging(b.buildCampCommand(cfg)),
	}
}

func (b *Bot) withCommandLogging(cmd commands.Command) commands.Command {
	var autoFn func(
		s *discordgo.Session,
		i *discordgo.InteractionCreate,
		focused *discordgo.ApplicationCommandInteractionDataOption,
	)

	if ac, ok := cmd.(commands.AutoCompleteCommand); ok {
		autoFn = ac.HandleAutoCompleteCommand
	}

	return commands.CommandFunc{
		NameFn:        cmd.Name,
		DescriptionFn: cmd.Description,
		SlashDefFn:    cmd.SlashDef,
		HandleFn: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			data := i.ApplicationCommandData()

			subName := ""
			if len(data.Options) > 0 &&
				data.Options[0].Type == discordgo.ApplicationCommandOptionSubCommand {
				subName = data.Options[0].Name
			}

			b.logCommandUsage(s, i, data.Name, subName)
			cmd.HandleInteraction(s, i)
		},
		AutoFn: autoFn,
	}
}
