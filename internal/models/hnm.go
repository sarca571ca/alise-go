package models

import (
	"fmt"
	"slices"
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
	Emoji          string
	Info           string
	Note           string
}

type HNMTimer struct {
	HNM         HNM
	Mod         string
	LastKill    time.Time
	DaysSinceHQ int
}

var HNMs = map[string]HNM{
	"fafnir": {
		ID:             "fafnir",
		Name:           "Fafnir",
		BaseRespawn:    22 * time.Hour,
		WindowInterval: 10 * time.Minute,
		WindowCount:    7,
		HQName:         "Nidhogg",
		Emoji:          ":dragon_face:",
		Note:           basicNote(),
	},
	"adamantoise": {
		ID:             "adamantoise",
		Name:           "Adamantoise",
		BaseRespawn:    22 * time.Hour,
		WindowInterval: 10 * time.Minute,
		WindowCount:    7,
		HQName:         "Aspidochelone",
		Emoji:          ":turtle:",
		Note:           basicNote(),
	},
	"behemoth": {
		ID:             "behemoth",
		Name:           "Behemoth",
		BaseRespawn:    22 * time.Hour,
		WindowInterval: 10 * time.Minute,
		WindowCount:    7,
		HQName:         "King Behemoth",
		Emoji:          ":zap:",
		Note:           basicNote(),
	},
	"tiamat": {
		ID:             "tiamat",
		Name:           "Tiamat",
		BaseRespawn:    84 * time.Hour,
		WindowInterval: 60 * time.Minute,
		WindowCount:    25,
		HQName:         "",
		Emoji:          ":fire::chicken::fire:",
		Note:           gwNote(),
	},
	"jormungand": {
		ID:             "jorm",
		Name:           "Jormungand",
		BaseRespawn:    84 * time.Hour,
		WindowInterval: 60 * time.Minute,
		WindowCount:    25,
		HQName:         "",
		Emoji:          ":ice_cube::chicken::ice_cube:",
		Note:           gwNote(),
	},
	"vrtra": {
		ID:             "vrtra",
		Name:           "Vrtra",
		BaseRespawn:    84 * time.Hour,
		WindowInterval: 60 * time.Minute,
		WindowCount:    25,
		HQName:         "",
		Emoji:          ":skull::chicken::skull:",
		Note:           gwNote(),
	},
	"simurgh": {
		ID:             "simurgh",
		Name:           "Simurgh",
		BaseRespawn:    22 * time.Hour,
		WindowInterval: 10 * time.Minute,
		WindowCount:    7,
		HQName:         "",
		Emoji:          ":bird:",
		Note:           basicNote(),
	},
	"ka": {
		ID:             "ka",
		Name:           "King Arthro",
		BaseRespawn:    22 * time.Hour,
		WindowInterval: 10 * time.Minute,
		WindowCount:    7,
		HQName:         "",
		Emoji:          ":crab:",
		Note:           basicNote(),
	},
	"bloodsucker": {
		ID:             "bloodsucker",
		Name:           "Bloodsucker",
		BaseRespawn:    71 * time.Hour,
		WindowInterval: 0 * time.Minute,
		WindowCount:    1,
		HQName:         "",
		Emoji:          ":drop_of_blood:",
		Note:           basicNote(),
	},
	"shiki": {
		ID:             "shiki",
		Name:           "Shikigami Weapon",
		BaseRespawn:    21 * time.Hour,
		WindowInterval: 0 * time.Minute,
		WindowCount:    1,
		HQName:         "",
		Emoji:          ":japanese_ogre:",
		Note:           basicNote(),
	},
	"kv": {
		ID:             "kv",
		Name:           "King Vinegarroon",
		BaseRespawn:    21 * time.Hour,
		WindowInterval: 0 * time.Minute,
		WindowCount:    1,
		HQName:         "",
		Emoji:          ":scorpion:",
		Note:           kvNote(),
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
		Mod:         h.Mod,
		LastKill:    pt,
		DaysSinceHQ: h.Day,
	}, nil

}

func CanHQ(hnmName string) bool {
	return slices.Contains([]string{"Fafnir", "Adamantoise", "Behemoth"}, hnmName)
}

func basicNote() string {
	return fmt.Sprintf(
		"- Channel will be open for 5-Minutes after pop/last window.\n" +
			"- Channel is moved to Awaiting Processing category.\n" +
			"- Late x-in's (within reason) or corrections to your camp status can be made after its moved.\n" +
			"- DO NOT X-IN before arriving to camp. This means in position and buffed.",
	)
}

func gwNote() string {
	return fmt.Sprintf(
		"- A valid hold party must be present for dkp.\n" +
			"- Conditions for valid hold party are: Tank (w/ Resist Set), BRD, WHM, 2 Sleeps)\n" +
			"- Windows will be opened 5-Minutes prior to window and closed 1-Minute after window.\n" +
			"- Late x-in's won't be allowed due to the nature of this camp.",
	)
}

func kvNote() string {
	return fmt.Sprintf(
		"- x         - used when you are at kv with the window open prior to pop\n" +
			"- x-pop     - used when you are present when KV pops and we do NOT claim\n" +
			"- x-claim   - used when you are present when KV pops and we DO claim\n" +
			"- x-kill    - used when you are present for the kill of KV\n\n" +
			"x-pop and x-claim are mutually exclusive",
	)
}
