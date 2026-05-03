package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func parseAutoComplete(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	partial string,
	choices []*discordgo.ApplicationCommandOptionChoice,
) {

	var matches []*discordgo.ApplicationCommandOptionChoice

	for c := range choices {
		name := strings.ToLower(choices[c].Name)
		val, _ := choices[c].Value.(string)
		value := strings.ToLower(val)
		if strings.HasPrefix(name, partial) || strings.HasPrefix(value, partial) {
			matches = append(matches, choices[c])
		}
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: matches,
		},
	})
}

func resolveHNM(s *discordgo.Session, i *discordgo.InteractionCreate) (string, bool) {
	data := i.ApplicationCommandData()

	var hnmAliases = map[string]string{
		"fafnir":           "fafnir",
		"faf":              "fafnir",
		"adamantoise":      "adamantoise",
		"ada":              "adamantoise",
		"behemoth":         "behemoth",
		"beh":              "behemoth",
		"behe":             "behemoth",
		"tiamat":           "tiamat",
		"tia":              "tiamat",
		"jormundgand":      "jormundgand",
		"jor":              "jormundgand",
		"jorm":             "jormundgand",
		"vrtra":            "vrtra",
		"vrt":              "vrtra",
		"ka":               "ka",
		"king arthro":      "ka",
		"kv":               "kv",
		"king vinegarron":  "kv",
		"shiki":            "shiki",
		"shi":              "shiki",
		"shikigami weapon": "shiki",
		"bloodsucker":      "bloodsucker",
		"bs":               "bloodsucker",
		"sim":              "simurgh",
		"simurgh":          "simurgh",
	}

	hnmOpt := data.GetOption("name")
	hnmKey := strings.TrimSpace(strings.ToLower(hnmOpt.Value.(string)))
	hnm, ok := hnmAliases[hnmKey]

	if !ok {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid input for HNM Name.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return "", false
	}
	return hnm, true
}

func resolveHNMMod(s *discordgo.Session, i *discordgo.InteractionCreate) (string, bool) {
	data := i.ApplicationCommandData()

	var modAliases = map[string]string{
		"normal":      "n",
		"approximate": "a",
		"double dkp":  "d",
		"triple dkp":  "t",
		"n":           "n",
		"a":           "a",
		"d":           "d",
		"t":           "t",
	}

	modOpt := data.GetOption("mod")
	var modKey string
	if modOpt == nil {
		modKey = "normal"
	} else {
		modKey = strings.TrimSpace(strings.ToLower(modOpt.Value.(string)))
	}
	mod, ok := modAliases[modKey]
	if !ok {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid input for HNM Mod.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return "", false
	}
	return mod, true
}

func getNameOption(sub *discordgo.ApplicationCommandInteractionDataOption) string {
	for _, opt := range sub.Options {
		if opt.Name == "name" {
			return opt.StringValue()
		}
	}
	return ""
}

func getQualityOption(sub *discordgo.ApplicationCommandInteractionDataOption) string {
	for _, opt := range sub.Options {
		if opt.Name == "quality" {
			return opt.StringValue()
		}
	}
	return ""
}

func getWindowOption(sub *discordgo.ApplicationCommandInteractionDataOption) int {
	for _, opt := range sub.Options {
		if opt.Name == "window" {
			return int(opt.IntValue())
		}
	}
	return 0
}
