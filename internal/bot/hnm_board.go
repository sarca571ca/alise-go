package bot

import (
	"alise-go/internal/data"
	"alise-go/internal/models"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func buildHNMTimerBoardEmbed(timers []models.HNMTimer) *discordgo.MessageEmbed {
	color := 0x760eed
	if len(timers) == 0 {
		return &discordgo.MessageEmbed{
			Title:       "HNM Camp Timers",
			Description: "No timers set for this channel yet.",
			Color:       color,
		}
	}

	type timerRow struct {
		timer   models.HNMTimer
		windows models.HNMTimerWindows
	}

	rows := make([]timerRow, 0, len(timers))
	for _, t := range timers {
		rows = append(rows, timerRow{
			timer:   t,
			windows: models.BuildHNMTimerWindows(t),
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].windows.NextRespawn.Before(rows[j].windows.NextRespawn)
	})

	fields := make([]*discordgo.MessageEmbedField, 0, len(rows))

	for _, row := range rows {
		t := row.timer
		w := row.windows

		name := models.BuildHNMTimerName(t)
		value := fmt.Sprintf(
			"<t:%d:T> <t:%d:R>",
			w.NextRespawn.Unix(),
			w.NextRespawn.Unix(),
		)

		fields = append(fields, &discordgo.MessageEmbedField{
			Value:  name + " " + value,
			Inline: false,
		})
	}

	return &discordgo.MessageEmbed{
		Title:  "HNM Camp Timers",
		Color:  color,
		Fields: fields,
	}
}

func (b *Bot) updateHNMTimerBoard(guildID string, timers []models.HNMTimer) error {
	channelID := b.cfg.Channels.HNMTimes
	if channelID == "" {
		return nil
	}

	embed := buildHNMTimerBoardEmbed(timers)

	board, found, err := b.store.GetHNMTimerBoard(guildID, channelID)
	if err != nil {
		return err
	}

	if !found {
		msg, err := b.dg.ChannelMessageSendEmbed(channelID, embed)
		if err != nil {
			return err
		}
		return b.store.UpsertHNMTimerBoard(data.HNMTimerBoard{
			GuildID:   guildID,
			ChannelID: channelID,
			MessageID: msg.ID,
		})
	}

	_, err = b.dg.ChannelMessageEditEmbed(channelID, board.MessageID, embed)
	return err
}

func (b *Bot) handleHNMCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := i.ApplicationCommandData().Options

	var input models.HNMInput
	for _, opt := range opts {
		switch opt.Name {
		case "name":
			input.Name = strings.ToLower(opt.StringValue())
		case "mod":
			input.Mod = opt.StringValue()
		case "day":
			input.Day = int(opt.IntValue())
		case "date":
			input.Date = int(opt.IntValue())
		case "time":
			input.Time = int(opt.IntValue())
		}
	}

	now := time.Now()

	timer, err := models.ParseHNMCommandInput(&input, now)
	if err != nil {
		respondEphemeral(s, i, "Invalid HNM input: "+err.Error())
		return
	}

	guildID := i.GuildID
	channelID := i.ChannelID // This needs to change to a specified channel

	rec := data.NewRecordFromHNMTimer(guildID, channelID, timer)

	if _, err := b.store.UpsertHNMTimerRecord(rec); err != nil {
		respondEphemeral(s, i, "Failed to load timers")
		return
	}

	recs, err := b.store.ListHNMTimerRecords(guildID, channelID)
	if err != nil {
		respondEphemeral(s, i, "Failed to load timers")
		return
	}

	var timers []models.HNMTimer
	for _, r := range recs {
		hnm, ok := models.GetHNM(r.HNMID)
		if !ok {
			continue
		}
		timers = append(timers, data.NewTimerFromRecord(r, hnm))
	}

	content := formatHNMTimersPlain(timers)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

func formatHNMTimersPlain(timers []models.HNMTimer) string {
	if len(timers) == 0 {
		return "No HNM timers set for this channel yet."
	}

	var sb strings.Builder
	sb.WriteString("Current HNM timers:\n")

	for _, t := range timers {
		w := models.BuildHNMTimerWindows(t)
		fmt.Fprintf(&sb,
			"- %s: last kill <t:%d:R>, next respawn <t:%d:R>\n",
			t.HNM.Name,
			t.LastKill.Unix(),
			w.NextRespawn.Unix(),
		)
	}

	return sb.String()
}

func formatHNMTimerPlain(timer models.HNMTimer) string {
	var sb strings.Builder

	w := models.BuildHNMTimerWindows(timer)
	fmt.Fprintf(&sb,
		"- %s: last kill <t:%d:R>, next respawn <t:%d:R>\n",
		timer.HNM.Name,
		timer.LastKill.Unix(),
		w.NextRespawn.Unix(),
	)

	return sb.String()
}
