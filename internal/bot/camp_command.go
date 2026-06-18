package bot

import (
	"alise-go/internal/commands"
	"alise-go/internal/config"
	"alise-go/internal/data"
	"alise-go/internal/formatting"
	"alise-go/internal/models"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) buildCampCommand(cfg config.Config) commands.Command {
	popCommand := commands.CampCommand{
		Pop: func(s *discordgo.Session, i *discordgo.InteractionCreate, lsName, quality string) {
			ch, err := s.Channel(i.ChannelID)
			if err != nil {
				respondEphemeral(s, i, "Failed to load channel.")
				return
			}

			if lsName == "" {
				respondEphemeral(s, i, "You must provide a linkshell name.")
				return
			}

			camp, err := b.camp.Pop(i.GuildID, lsName, quality, ch)
			if err != nil {
				respondEphemeral(s, i, err.Error())
				return
			}

			hnm, ok := models.GetHNM(camp.HNMID)
			if !ok {
				respondEphemeral(s, i, "Claim recorded, but failed to resolve HNM name.")
				go b.hnm.MoveCampAfterDelay(camp.ChannelID, 5*time.Minute)
				return
			}

			respondEphemeral(
				s,
				i,
				fmt.Sprintf("%s claim recorded for %s.", lsName, hnm.Name),
			)

			if quality != "" {
				_, _ = b.dg.ChannelMessageSend(i.ChannelID, formatting.FormatWindowHeading(fmt.Sprintf("Pop | %s | %s", lsName, quality)))
			} else {
				_, _ = b.dg.ChannelMessageSend(i.ChannelID, formatting.FormatWindowHeading(fmt.Sprintf("Pop | %s", lsName)))
			}
			_, _ = b.dg.ChannelMessageSend(i.ChannelID, "Moving channel to awaiting-processing in 5 minutes.")
			go b.hnm.MoveCampAfterDelay(camp.ChannelID, 5*time.Minute)

			// NOTE: Uncomment lines to have a dedicated channel for LeaderBoards
			// recs, err := b.store.ListLinkshellRecords(cfg.GuildID)
			// if err != nil {
			// 	respondEphemeral(s, i, "Failed to load linkshell recs")
			// }
			//
			// linkshells := b.store.NewLinkshellListFromRecords(recs)
			//
			// if err := b.updateLinkshellLeaderBoard(cfg.GuildID, linkshells); err != nil {
			// 	respondEphemeral(s, i, "Failed to update linkshell leader boards")
			// 	return
			// }
		},
		Dead: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			now := time.Now()
			camp, ok, err := b.store.GetHNMCampChannelByChannelID(i.GuildID, i.ChannelID)
			if err != nil {
				respondEphemeral(s, i, "failed to load camp")
				return
			}
			if !ok {
				respondEphemeral(s, i, "no camp record found for this channel")
				return
			}

			if !camp.IsSpawned {
				respondEphemeral(s, i, "/camp pop must be used first, manual ToD entry now required")
				return
			}

			hnm, ok := models.GetHNM(camp.HNMID)
			if !ok {
				return
			}

			timer := models.HNMTimer{
				HNM:      hnm,
				Mod:      "n",
				LastKill: now,
			}

			if hnm.HQName != "" {
				if camp.IsHQ {
					timer.DaysSinceHQ = 1
				} else {
					timer.DaysSinceHQ = camp.DaysSinceHQ + 1
				}
			}

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
		Open: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			camp, ok, err := b.store.GetHNMCampChannelByChannelID(i.GuildID, i.ChannelID)
			if err != nil {
				respondEphemeral(s, i, "failed to load camp")
				return
			}
			if !ok {
				respondEphemeral(s, i, "no camp record found for this channel")
				return
			}
			if !camp.IsClosed {
				respondEphemeral(s, i, "camp is not closed")
				return
			}
			camp.IsClosed = false
			if _, err := b.store.UpsertHNMCampChannel(camp); err != nil {
				respondEphemeral(s, i, "failed to update camp")
				return
			}

			respondEphemeral(s, i, "camp is now open")
			_, _ = b.dg.ChannelMessageSend(i.ChannelID, formatting.FormatWindowHeading("Open"))
		},
		Close: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			camp, ok, err := b.store.GetHNMCampChannelByChannelID(i.GuildID, i.ChannelID)
			if err != nil {
				respondEphemeral(s, i, "failed to load camp")
				return
			}
			if !ok {
				respondEphemeral(s, i, "no camp record found for this channel")
				return
			}
			if camp.IsClosed {
				respondEphemeral(s, i, "camp is not open")
				return
			}
			camp.IsClosed = true
			if _, err := b.store.UpsertHNMCampChannel(camp); err != nil {
				respondEphemeral(s, i, "failed to update camp")
				return
			}

			respondEphemeral(s, i, "camp is now closed use /camp open to re-open")
			_, _ = b.dg.ChannelMessageSend(i.ChannelID, formatting.FormatWindowHeading("Closed"))
		},
		Enrage: func(s *discordgo.Session, i *discordgo.InteractionCreate, window int) {
			if window == 0 {
				respondEphemeral(s, i, "enrage must have a window number of atleast 1")
				return
			}

			ch, err := s.Channel(i.ChannelID)
			if err != nil {
				respondEphemeral(s, i, "Failed to load channel.")
				return
			}

			camp, ok, err := b.store.GetHNMCampChannelByChannelID(i.GuildID, i.ChannelID)
			if err != nil {
				respondEphemeral(s, i, "failed to load camp")
				return
			}
			if !ok {
				respondEphemeral(s, i, "no camp record found for this channel")
				return
			}

			camp.MoveScheduled = false
			camp.IsEnraged = true

			if ch.ParentID != cfg.Categories.HNMCategoryID {
				guildID := cfg.GuildID
				targetParent := cfg.Categories.HNMCategoryID
				if guildID == "" || targetParent == "" {
					return
				}

				_, _ = b.dg.ChannelEdit(i.ChannelID, &discordgo.ChannelEdit{
					ParentID: targetParent,
				})
			}

			if _, err := b.store.UpsertHNMCampChannel(camp); err != nil {
				respondEphemeral(s, i, "failed to update camp")
				return
			}

			respondEphemeral(s, i, "camp is now enraged")
			_, _ = b.dg.ChannelMessageSend(i.ChannelID, formatting.FormatWindowHeading(fmt.Sprintf("Enrage %v", window)))
			go b.camp.EnrageWindow(i.ChannelID, window, 5*time.Minute, camp)
			go b.hnm.MoveCampAfterDelay(i.ChannelID, 10*time.Minute)
		},
		Continue: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			ch, err := s.Channel(i.ChannelID)
			if err != nil {
				respondEphemeral(s, i, "Failed to load channel.")
				return
			}

			camp, ok, err := b.store.GetHNMCampChannelByChannelID(i.GuildID, i.ChannelID)
			if err != nil {
				respondEphemeral(s, i, "failed to load camp")
				return
			}
			if !ok {
				respondEphemeral(s, i, "no camp record found for this channel")
				return
			}

			camp.MoveScheduled = false
			camp.IsEnraged = false

			if ch.ParentID != cfg.Categories.HNMCategoryID {
				guildID := cfg.GuildID
				targetParent := cfg.Categories.HNMCategoryID
				if guildID == "" || targetParent == "" {
					return
				}

				_, _ = b.dg.ChannelEdit(i.ChannelID, &discordgo.ChannelEdit{
					ParentID: targetParent,
				})
			}

			if _, err := b.store.UpsertHNMCampChannel(camp); err != nil {
				respondEphemeral(s, i, "failed to update camp")
				return
			}

			respondEphemeral(s, i, "camp is now extended, use /camp stable to close")
			_, _ = b.dg.ChannelMessageSend(i.ChannelID, formatting.FormatWindowHeading("Camp Extended"))
		},
		Stable: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			camp, ok, err := b.store.GetHNMCampChannelByChannelID(i.GuildID, i.ChannelID)
			if err != nil {
				respondEphemeral(s, i, "failed to load camp")
				return
			}
			if !ok {
				respondEphemeral(s, i, "no camp record found for this channel")
				return
			}

			camp.MoveScheduled = true
			camp.IsEnraged = false

			if _, err := b.store.UpsertHNMCampChannel(camp); err != nil {
				respondEphemeral(s, i, "failed to update camp")
				return
			}

			respondEphemeral(s, i, "camp is now stable")
			go b.hnm.MoveCampAfterDelay(camp.ChannelID, 5*time.Minute)
		},
		CTA: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// 1. create a new channel using the same channel name just adding -cta to the end
			ch, err := s.Channel(i.ChannelID)
			if err != nil {
				respondEphemeral(s, i, "Failed to load channel.")
				return
			}

			channel, err := b.dg.GuildChannelCreateComplex(cfg.GuildID, discordgo.GuildChannelCreateData{
				Name:     ch.Name + "-cta",
				Type:     discordgo.ChannelTypeGuildText,
				ParentID: cfg.Categories.HNMCategoryID,
			})
			if err != nil {
				respondEphemeral(s, i, err.Error())
				return
			}
			if _, err := b.dg.ChannelMessageSend(channel.ID, formatting.FormatWindowHeading("Call to Arms")); err != nil {
				respondEphemeral(s, i, err.Error())
				return
			}
		},
	}
	return popCommand
}
