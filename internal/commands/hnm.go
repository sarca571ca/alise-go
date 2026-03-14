package commands

import (
	"log"
	"strings"
	"time"

	"alise-go/internal/models"

	"github.com/bwmarrin/discordgo"
)

type HNMCommand struct{}

func (HNMCommand) Name() string        { return "hnm" }
func (HNMCommand) Description() string { return "Make a camp timer." }
func (HNMCommand) SlashDef() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "hnm",
		Description: "Make a camp timer",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:         "name",
				Description:  "HNM name (e.g. Fafnir, Adamantoise)",
				Type:         discordgo.ApplicationCommandOptionString,
				Required:     true,
				Autocomplete: true,
			},
			{
				Name:         "time",
				Description:  "Time of the last kill. (e.g. 102356, 224503)",
				Type:         discordgo.ApplicationCommandOptionInteger,
				Required:     true,
				Autocomplete: false,
			},
			{
				Name:         "date",
				Description:  "Date of the last kill. (e.g. 20260130, 20260516)",
				Type:         discordgo.ApplicationCommandOptionInteger,
				Required:     false,
				Autocomplete: false,
			},
			{
				Name:         "day",
				Description:  "How many days since HQ. Will be 0 if HNM doesn't HQ.(e.g. 0, 1, 2)",
				Type:         discordgo.ApplicationCommandOptionInteger,
				Required:     false,
				Autocomplete: false,
			},
			{
				Name:         "mod",
				Description:  "Camp modifier (e.g. n[normal], d[double dkp], t[triple dkps], a[approximate])",
				Type:         discordgo.ApplicationCommandOptionString,
				Required:     false,
				Autocomplete: true,
			},
		},
	}
}

func (HNMCommand) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	// HNM Name
	hnm, ok := resolveHNM(s, i)
	if !ok {
		return
	}

	// HNM Mod
	mod, ok := resolveMod(s, i)
	if !ok {
		return
	}

	// HNM HQ Day
	dayOpt := data.GetOption("day")
	var day int
	if dayOpt == nil {
		day = 0
	} else {
		day = int(dayOpt.IntValue())
	}

	// HNM Date
	dateOpt := data.GetOption("date")
	var date int
	if dateOpt == nil {
		date = 0
	} else {
		date = int(dateOpt.IntValue())
	}

	// HNM Time
	timeOpt := data.GetOption("time")
	timestamp := int(timeOpt.IntValue())

	hnmInput := models.HNMInput{
		Name: hnm,
		Mod:  mod,
		Day:  day,
		Date: date,
		Time: timestamp,
	}

	hnmTimer, err := models.ParseHNMCommandInput(&hnmInput, time.Now())
	if err != nil {
		return
	}
	hnmTimerWindows := models.BuildHNMTimerWindows(hnmTimer)
	log.Println(hnmTimerWindows)
}

func (HNMCommand) HandleAutoCompleteCommand(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	focused *discordgo.ApplicationCommandInteractionDataOption,
) {
	partial := strings.ToLower(focused.StringValue())

	switch focused.Name {
	case "name":
		choices := []*discordgo.ApplicationCommandOptionChoice{
			{Name: "Fafnir", Value: "fafnir"},
			{Name: "Adamantoise", Value: "adamantoise"},
			{Name: "Behemoth", Value: "behemoth"},
			{Name: "Tiamat", Value: "tiamat"},
			{Name: "Jormundgand", Value: "jormundgand"},
			{Name: "Vrtra", Value: "vrtra"},
			{Name: "King Arthro", Value: "ka"},
			{Name: "King Vinegarron", Value: "kv"},
			{Name: "Simurgh", Value: "simurgh"},
			{Name: "Shikigami Weapon", Value: "shiki"},
			{Name: "Bloodsucker", Value: "bloodsucker"},
		}

		parseAutoComplete(s, i, partial, choices)

	case "mod":
		choices := []*discordgo.ApplicationCommandOptionChoice{
			{Name: "Normal", Value: "n"},
			{Name: "Approximate", Value: "a"},
			{Name: "Double DKP", Value: "d"},
			{Name: "Triple DKP", Value: "t"},
		}

		parseAutoComplete(s, i, partial, choices)
	}

}
