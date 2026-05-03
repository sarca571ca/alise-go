package models

import (
	"fmt"
	"time"
)

type HNMTimerWindows struct {
	NextRespawn time.Time
	Windows     []time.Time
}

func BuildHNMTimerWindows(t HNMTimer) HNMTimerWindows {
	next := t.LastKill.Add(t.HNM.BaseRespawn)

	wins := make([]time.Time, t.HNM.WindowCount)
	for i := 0; i < t.HNM.WindowCount; i++ {
		wins[i] = next.Add(time.Duration(i) * t.HNM.WindowInterval)
	}

	return HNMTimerWindows{
		NextRespawn: next,
		Windows:     wins,
	}
}

func BuildHNMTimerName(t HNMTimer) string {
	name := t.HNM.Name
	mod := getHNMMod(t.Mod)

	if !CanHQ(t.HNM.Name) {
		return fmt.Sprintf("%s %s%s%s:", name, mod, t.HNM.Emoji, mod)
	}

	if t.DaysSinceHQ > 3 {
		return fmt.Sprintf(
			"**%s/%s** :rotating_light:%s%s%s(**%d**):",
			name,
			t.HNM.HQName,
			mod,
			t.HNM.Emoji,
			mod,
			t.DaysSinceHQ,
		)
	}

	return fmt.Sprintf("%s %s%s%s(**%d**):", name, mod, t.HNM.Emoji, mod, t.DaysSinceHQ)
}

func getHNMMod(mod string) string {
	mods := map[string]string{
		"":  "",
		"n": "",
		"a": ":grey_question:",
		"d": ":moneybag:",
		"t": ":gem:",
	}
	return mods[mod]
}
