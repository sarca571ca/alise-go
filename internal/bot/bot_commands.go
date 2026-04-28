package bot

import (
	"alise-go/internal/commands"
	"alise-go/internal/config"
	"alise-go/internal/data"
	"alise-go/internal/models"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) buildCommands(cfg config.Config) []commands.Command {
	return []commands.Command{
		b.withCommandLogging(b.buildHnmCommand(cfg)),
		b.withCommandLogging(b.buildLinkshellCommand(cfg)),
	}
}

func (b *Bot) buildHnmCommand(cfg config.Config) commands.Command {
	hnmCmd := commands.HNMCommand{
		OnParsed: func(s *discordgo.Session, i *discordgo.InteractionCreate, timer models.HNMTimer) {
			guildID := i.GuildID
			channelID := cfg.Channels.HNMTimes

			rec := data.NewRecordFromHNMTimer(guildID, channelID, timer)

			if _, err := b.store.UpsertHNMTimerRecord(rec); err != nil {
				respondEphemeral(s, i, "Failed to save timer")
				return
			}

			recs, err := b.store.ListHNMTimerRecords(guildID, channelID)
			if err != nil {
				respondEphemeral(s, i, "Failed to load timers")
			}

			var timers []models.HNMTimer
			for _, r := range recs {
				hnm, ok := models.GetHNM(r.HNMID)
				if !ok {
					continue
				}
				timers = append(timers, data.NewTimerFromRecord(r, hnm))
			}

			if err := b.updateHNMTimerBoard(guildID, timers); err != nil {
				respondEphemeral(s, i, "Failed to update timer board")
				return
			}

			content := formatHNMTimersPlain(timers)

			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
				},
			},
			)
		},
	}
	return hnmCmd
}

func (b *Bot) buildLinkshellCommand(cfg config.Config) commands.Command {
	linkshellCmd := commands.LinkshellCommand{
		// TODO: I think we'll need to add the handlers for each of the mods for this
		// command. I'm thinking i want to have a Linkshell Leaderboard like the timerboard.
		// I maybe getting ahead of myself a bit though. The pop command will update a
		// individual linkshell and hnm claim. It might be best to do a dialog box for the
		// update mod so you can adjust a value if its inputted incorrectly.
		List: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			guildID := i.GuildID
			recs, err := b.store.ListLinkshellRecords(guildID)
			if err != nil {
				respondEphemeral(s, i, "Failed to list linkshell records.")
			}

			msg := formatLinkshellListPlain(b.store.NewLinkshellListFromRecords(recs))
			respondEphemeral(s, i, msg)
		},
		Add: func(s *discordgo.Session, i *discordgo.InteractionCreate, ls string) {
			guildID := i.GuildID
			_, ok, err := b.store.GetLinkshellRecord(guildID, ls)
			if err != nil {
				respondEphemeral(s, i, err.Error())
				return
			}
			if ok {
				respondEphemeral(s, i, "The linkshell already exists.")
				return
			}
			rec := b.store.NewRecordFromLinkshellEntry(guildID, ls)
			if _, err := b.store.UpsertLinkshellRecord(rec); err != nil {
				respondEphemeral(s, i, "Failed to save Linkshell")
				return
			}

			var sb strings.Builder
			fmt.Fprintf(&sb, "Successfully added %s to the linkshells list.", ls)
			respondEphemeral(s, i, sb.String())

		},
		// TODO: Remove command for the Linkshell List
		Remove: func(s *discordgo.Session, i *discordgo.InteractionCreate, ls string) {
			guildID := i.GuildID
			_, ok, err := b.store.GetLinkshellRecord(guildID, ls)
			if err != nil {
				respondEphemeral(s, i, err.Error())
				return
			}
			if !ok {
				respondEphemeral(s, i, "The linkshell doesn't exist.")
				return
			}
			if err := b.store.DeleteLinkshellRecord(guildID, ls); err != nil {
				respondEphemeral(s, i, "Failed to remove Linkshell")
				return
			}

			var sb strings.Builder
			fmt.Fprintf(&sb, "Successfully removed %s to the linkshells list.", ls)
			respondEphemeral(s, i, sb.String())
		},
		// TODO: Update modal that self populates the the fields with current values
		Update: func(s *discordgo.Session, i *discordgo.InteractionCreate, ls string) {},
	}
	return linkshellCmd
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
