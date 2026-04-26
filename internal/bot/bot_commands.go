package bot

import (
	"alise-go/internal/commands"
	"alise-go/internal/config"
	"alise-go/internal/data"
	"alise-go/internal/models"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) buildCommands(cfg config.Config) []commands.Command {
	return []commands.Command{
		b.buildHnmCommand(cfg),
		b.buildLinkshellCommand(cfg),
	}
}

func (b *Bot) buildHnmCommand(cfg config.Config) commands.Command {
	hnmCmd := commands.HNMCommand{
		OnParsed: func(s *discordgo.Session, i *discordgo.InteractionCreate, timer models.HNMTimer) {
			guildID := i.GuildID
			channelID := cfg.Channels.HNMTimes

			rec := data.NewRecordFromHNMTimer(guildID, channelID, timer)

			if _, err := b.store.UpsertHNMTimerRecord(rec); err != nil {
				respondError(s, i, "Failed to save timer")
				return
			}

			recs, err := b.store.ListHNMTimerRecords(guildID, channelID)
			if err != nil {
				respondError(s, i, "Failed to load timers")
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
				respondError(s, i, "Failed to update timer board")
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
				respondError(s, i, "Failed to list linkshell records.")
			}

			linkshellList := b.store.NewLinkshellListFromRecords(recs)
			respondWithLinkshellList(s, i, linkshellList)
		},
		Add: func(s *discordgo.Session, i *discordgo.InteractionCreate, ls string) {
			guildID := i.GuildID
			_, ok, err := b.store.GetLinkshellRecord(guildID, ls)
			if err != nil {
				respondError(s, i, err.Error())
				return
			}
			if ok {
				respondError(s, i, "The linkshell already exists.")
				return
			}
			rec := b.store.NewRecordFromLinkshellAdd(guildID, ls)
			if _, err := b.store.UpsertLinkshellRecord(rec); err != nil {
				respondError(s, i, "Failed to save Linkshell")
				return
			}

		},
		Remove: func(s *discordgo.Session, i *discordgo.InteractionCreate, ls string) {},
		Update: func(s *discordgo.Session, i *discordgo.InteractionCreate, ls string) {},
	}
	return linkshellCmd
}
