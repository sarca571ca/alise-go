package services

import (
	"alise-go/internal/data"
	"alise-go/internal/models"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func shortMonth(t time.Time) string {
	return strings.ToLower(t.Format("Jan"))
}

func hnmShortCode(hnm models.HNM) string {
	switch hnm.ID {
	case "behemoth":
		return "beh"
	case "adamantoise":
		return "ada"
	case "fafnir":
		return "faf"
	case "tiamat":
		return "tia"
	case "jorm":
		return "jor"
	case "vrtra":
		return "vrt"
	case "simurgh":
		return "sim"
	case "ka":
		return "ka"
	case "bloodsucker":
		return "bs"
	case "shiki":
		return "shi"
	case "kv":
		return "kv"
	default:
		return hnm.ID
	}
}

func campNameAndSeq(store *data.Store, guildID string, timer models.HNMTimer) (string, int, error) {
	t := timer.LastKill.Add(timer.HNM.BaseRespawn)

	mon := shortMonth(t)
	day := t.Day()

	short := hnmShortCode(timer.HNM)

	hqSuffix := ""
	if timer.HNM.HQName != "" {
		hqSuffix = fmt.Sprintf("%d", timer.DaysSinceHQ+1)
	}

	existing, err := store.ListHNMCampChannelsForDay(guildID, timer.HNM.ID, t)
	if err != nil {
		return "", 0, err
	}

	seq := 0
	if len(existing) > 0 {
		seq = 1
	}

	seqSuffix := ""
	if seq > 0 {
		seqSuffix = fmt.Sprintf("%d", seq)
	}

	name := fmt.Sprintf("%s%d-%s%s%s", mon, day, short, hqSuffix, seqSuffix)
	return name, seq, nil
}

func currentWindowIndex(now time.Time, wins models.HNMTimerWindows) int {
	if len(wins.Windows) == 0 {
		return 0
	}

	if now.Before(wins.Windows[0]) || now.After(wins.Windows[len(wins.Windows)-1]) {
		return 0
	}

	for idx, w := range wins.Windows {
		nextIdx := idx + 1
		if now.Before(w) {
			return idx
		}
		if nextIdx == len(wins.Windows) && !now.Before(w) {
			return nextIdx
		}
	}

	return 0
}

func (s *HNMService) moveCampAfterDelay(channelID string, delay time.Duration) {
	time.Sleep(delay)

	// Move to AwaitingProcessing category.
	guildID := s.cfg.GuildID
	targetParent := s.cfg.Categories.AwaitingProcessingID
	if guildID == "" || targetParent == "" {
		return
	}

	_, _ = s.dg.ChannelEdit(channelID, &discordgo.ChannelEdit{
		ParentID: targetParent,
	})
}
