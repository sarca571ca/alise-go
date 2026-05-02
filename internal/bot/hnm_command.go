package bot

import (
	"alise-go/internal/commands"
	"alise-go/internal/config"
	"alise-go/internal/data"
	"alise-go/internal/models"

	"github.com/bwmarrin/discordgo"
)

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

			content := formatHNMTimerPlain(timer)

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
