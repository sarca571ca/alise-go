// package bot
//
// import (
// 	"alise-go/internal/data"
// 	"alise-go/internal/models"
// 	"fmt"
// 	"strings"
// 	"time"
//
// 	"github.com/bwmarrin/discordgo"
// )
//
// func respondError(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
// 	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
// 		Type: discordgo.InteractionResponseChannelMessageWithSource,
// 		Data: &discordgo.InteractionResponseData{
// 			Content: msg,
// 			Flags:   discordgo.MessageFlagsEphemeral, // only the user sees it
// 		},
// 	})
// }
//
// func handleHNMCommand(s *discordgo.Session, i *discordgo.InteractionCreate, store *data.Store) {
// 	opts := i.ApplicationCommandData().Options
//
// 	var input models.HNMInput
// 	for _, opt := range opts {
// 		switch opt.Name {
// 		case "name":
// 			input.Name = strings.ToLower(opt.StringValue())
// 		case "mod":
// 			input.Mod = opt.StringValue()
// 		case "day":
// 			input.Day = int(opt.IntValue())
// 		case "date":
// 			input.Date = int(opt.IntValue())
// 		case "time":
// 			input.Time = int(opt.IntValue())
// 		}
// 	}
//
// 	now := time.Now()
//
// 	timer, err := models.ParseHNMCommandInput(&input, now)
// 	if err != nil {
// 		respondError(s, i, "Invalid HNM input: "+err.Error())
// 		return
// 	}
//
// 	guildID := i.GuildID
// 	channelID := i.ChannelID // This needs to change to a specified channel
//
// 	rec := data.NewRecordFromHNMTimer(guildID, channelID, timer)
//
// 	if _, err := store.UpsertHNMTimerRecord(rec); err != nil {
// 		respondError(s, i, "Failed to load timers")
// 		return
// 	}
//
// 	recs, err := store.ListHNMTimerRecords(guildID, channelID)
// 	if err != nil {
// 		respondError(s, i, "Failed to load timers")
// 		return
// 	}
//
// 	var timers []models.HNMTimer
// 	for _, r := range recs {
// 		hnm, ok := models.GetHNM(r.HNMID)
// 		if !ok {
// 			continue
// 		}
// 		timers = append(timers, data.NewTimerFromRecord(r, hnm))
// 	}
//
// 	content := formatHNMTimersPlain(timers)
//
// 	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
// 		Type: discordgo.InteractionResponseChannelMessageWithSource,
// 		Data: &discordgo.InteractionResponseData{
// 			Content: content,
// 		},
// 	})
// }
//
// func formatHNMTimersPlain(timers []models.HNMTimer) string {
// 	if len(timers) == 0 {
// 		return "No HNM timers set for this channel yet."
// 	}
//
// 	var b strings.Builder
// 	b.WriteString("Current HNM timers:\n")
//
// 	for _, t := range timers {
// 		w := models.BuildHNMTimerWindows(t)
// 		fmt.Fprintf(&b,
// 			"- %s: last kill <t:%d:R>, next respawn <t:%d:R>\n",
// 			t.HNM.Name,
// 			t.LastKill.Unix(),
// 			w.NextRespawn.Unix(),
// 		)
// 	}
//
// 	return b.String()
// }
