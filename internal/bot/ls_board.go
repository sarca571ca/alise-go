package bot

import (
	"alise-go/internal/models"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func formatLinkshellListPlain(linkshells []models.Linkshell) string {
	// TODO: This will need more expansion/formating fixes but its a working concept for now.
	if len(linkshells) == 0 {
		return "No linkshells currently registered."
	}

	var b strings.Builder
	b.WriteString("Claims Leaderboards\n")
	b.WriteString("Linkshell\tFafnir(**Nidhogg**)\tAdamantoise(**Aspidochelone**)\tBehemoth(**King Behemoth**)\n")

	for _, ls := range linkshells {
		fmt.Fprintf(&b,
			"%s\t%v(**%v**)\t%v(**%v**)\t%v(**%v**)\n",
			ls.LinkshellName,
			ls.FafnirClaims,
			ls.NidhoggClaims,
			ls.AdamantoiseClaims,
			ls.AspidocheloneClaims,
			ls.BehemothClaims,
			ls.KingBehemothClaims,
		)
	}

	return b.String()
}

func respondWithLinkshellList(s *discordgo.Session, i *discordgo.InteractionCreate, linkshells []models.Linkshell) {
	msg := formatLinkshellListPlain(linkshells)
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
