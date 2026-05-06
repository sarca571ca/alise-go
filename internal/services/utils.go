package services

import (
	"alise-go/internal/data"
	"alise-go/internal/models"
	"strings"
	"time"
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

func shouldArchiveCamp(camp data.HNMCampChannel, timer data.HNMTimerRecord) bool {
	if !camp.MoveScheduled {
		return false
	}

	if !timer.LastKill.Equal(camp.LastKill) || timer.DaysSinceHQ != camp.DaysSinceHQ {
		return true
	}

	return false
}
