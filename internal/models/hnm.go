package models

import (
	"fmt"
	"time"
)

type HNMInput struct {
	Name string
	Mod  string
	Day  int
	Date int
	Time int
}

type HNM struct {
	ID             string
	Name           string
	BaseRespawn    time.Duration
	WindowInterval time.Duration
	WindowCount    int
	HQName         string
}

type HNMTimer struct {
	HNM         HNM
	LastKill    time.Time
	DaysSinceHQ int
}

var HNMs = map[string]HNM{
	"fafnir": {
		ID:             "fafnir",
		Name:           "Fafnir",
		BaseRespawn:    22 * time.Hour,
		WindowInterval: 10 * time.Second,
		WindowCount:    7,
		HQName:         "Nidhogg",
	},
	"adamantoise": {
		ID:             "adamantoise",
		Name:           "Adamantoise",
		BaseRespawn:    22 * time.Hour,
		WindowInterval: 10 * time.Minute,
		WindowCount:    7,
		HQName:         "Aspidchelone",
	},
	"behemoth": {
		ID:             "behemoth",
		Name:           "Behemoth",
		BaseRespawn:    22 * time.Hour,
		WindowInterval: 10 * time.Minute,
		WindowCount:    7,
		HQName:         "King Behemoth",
	},
	"tiamat": {
		ID:             "tiamat",
		Name:           "Tiamat",
		BaseRespawn:    84 * time.Hour,
		WindowInterval: 60 * time.Minute,
		WindowCount:    25,
		HQName:         "",
	},
	"jormungand": {
		ID:             "jorm",
		Name:           "Jormungand",
		BaseRespawn:    84 * time.Hour,
		WindowInterval: 60 * time.Minute,
		WindowCount:    25,
		HQName:         "",
	},
	"vrtra": {
		ID:             "vrtra",
		Name:           "Vrtra",
		BaseRespawn:    84 * time.Hour,
		WindowInterval: 60 * time.Minute,
		WindowCount:    25,
		HQName:         "",
	},
	"simurgh": {
		ID:             "simurgh",
		Name:           "Simurgh",
		BaseRespawn:    22 * time.Hour,
		WindowInterval: 10 * time.Minute,
		WindowCount:    7,
		HQName:         "",
	},
	"ka": {
		ID:             "ka",
		Name:           "King Arthro",
		BaseRespawn:    22 * time.Hour,
		WindowInterval: 10 * time.Minute,
		WindowCount:    7,
		HQName:         "",
	},
	"bloodsucker": {
		ID:             "bloodsucker",
		Name:           "Bloodsucker",
		BaseRespawn:    71 * time.Hour,
		WindowInterval: 0 * time.Minute,
		WindowCount:    1,
		HQName:         "",
	},
	"shiki": {
		ID:             "shiki",
		Name:           "Shikigami Weapon",
		BaseRespawn:    21 * time.Hour,
		WindowInterval: 0 * time.Minute,
		WindowCount:    1,
		HQName:         "",
	},
	"kv": {
		ID:             "kv",
		Name:           "King Vinegarroon",
		BaseRespawn:    21 * time.Hour,
		WindowInterval: 0 * time.Minute,
		WindowCount:    1,
		HQName:         "",
	},
}

func GetHNM(id string) (HNM, bool) {
	hnm, ok := HNMs[id]
	if !ok {
		return HNM{}, false
	}
	return hnm, true
}

func ParseHNMCommandInput(h *HNMInput, now time.Time) (HNMTimer, error) {
	dateProvided := h.Date != 0
	year, month, day := SplitYYYYMMDD(h.Date)
	hour, minute, second := SplitHHMMSS(h.Time)

	dt := DateTimeInput{
		Year:            year,
		Month:           month,
		Day:             day,
		Hour:            hour,
		Minute:          minute,
		Second:          second,
		HasExplicitDate: dateProvided,
	}

	pt := dt.Resolve(now)

	hnm, ok := GetHNM(h.Name)
	if !ok {
		return HNMTimer{}, fmt.Errorf("Unknown HNM %q", h.Name)
	}
	return HNMTimer{
		HNM:         hnm,
		LastKill:    pt,
		DaysSinceHQ: h.Day,
	}, nil

}
