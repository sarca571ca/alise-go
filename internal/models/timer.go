package models

import "time"

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
