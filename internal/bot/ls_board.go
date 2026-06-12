package bot

import (
	"alise-go/internal/models"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func buildLinkshellLeaderBoard(linkshells []models.Linkshell) *discordgo.MessageEmbed {
	color := 0x760eed
	if len(linkshells) == 0 {
		return &discordgo.MessageEmbed{
			Title:       "Linkshell Claims Leader Board",
			Description: "No Linkshells registered yet.",
			Color:       color,
		}
	}

	var fields []*discordgo.MessageEmbedField
	field := discordgo.MessageEmbedField{
		Name:  "",
		Value: formatLinkshellListPlain(linkshells),
	}
	fields = append(fields, &field)

	return &discordgo.MessageEmbed{
		Title:  "HNM Camp Timers",
		Color:  color,
		Fields: fields,
	}

}

// NOTE: Uncomment bellow lines to enable leaderboards
// func (b *Bot) updateLinkshellLeaderBoard(guildID string, linkshells []models.Linkshell) error {
// 	channelID := b.cfg.Channels.ClaimsLeaderBoard
// 	if channelID == "" {
// 		return nil
// 	}
//
// 	embed := buildLinkshellLeaderBoard(linkshells)
//
// 	board, found, err := b.store.GetLinkshellLeaderBoard(guildID, channelID)
// 	if err != nil {
// 		return err
// 	}
//
// 	if !found {
// 		msg, err := b.dg.ChannelMessageSendEmbed(channelID, embed)
// 		if err != nil {
// 			return err
// 		}
// 		return b.store.UpsertLinkshellLeaderBoard(data.LinkshellLeaderBoard{
// 			GuildID:   guildID,
// 			ChannelID: channelID,
// 			MessageID: msg.ID,
// 		})
// 	}
//
// 	_, err = b.dg.ChannelMessageEditEmbed(channelID, board.MessageID, embed)
// 	return err
//
// }

func formatLinkshellListPlain(linkshells []models.Linkshell) string {
	if len(linkshells) == 0 {
		return "No linkshells currently registered."
	}

	var sb strings.Builder
	sb.WriteString("```")
	sb.WriteString("Linkshell\tTotal\tHQ\tGW\tFaf(**Nid**)\tAda(**Asp**)\tBeh(**KB**)\tTia\tJorm\tVrt\tKV\tKA\tSim\tShiki\tBS\n")

	for _, ls := range linkshells {
		fmt.Fprintf(&sb,
			"%s\t%v\t%v\t%v\t%v(**%v**)\t%v(**%v**)\t%v(**%v**)\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n",
			ls.LinkshellName,
			totalClaims(ls),
			totalHQClaims(ls),
			totalGWClaims(ls),
			ls.FafnirClaims,
			ls.NidhoggClaims,
			ls.AdamantoiseClaims,
			ls.AspidocheloneClaims,
			ls.BehemothClaims,
			ls.KingBehemothClaims,
			ls.TiamatClaims,
			ls.JormungandClaims,
			ls.VrtraClaims,
			ls.KingVinegarroonClaims,
			ls.KingArthroClaims,
			ls.SimurghClaims,
			ls.ShikigamiWeaponClaims,
			ls.BloodsuckerClaims,
		)
	}

	sb.WriteString("```")

	return sb.String()
}

func totalClaims(linkshell models.Linkshell) int {
	return linkshell.AdamantoiseClaims +
		linkshell.AspidocheloneClaims +
		linkshell.FafnirClaims +
		linkshell.NidhoggClaims +
		linkshell.BehemothClaims +
		linkshell.KingBehemothClaims +
		linkshell.TiamatClaims +
		linkshell.JormungandClaims +
		linkshell.VrtraClaims +
		linkshell.KingVinegarroonClaims +
		linkshell.KingArthroClaims +
		linkshell.SimurghClaims +
		linkshell.ShikigamiWeaponClaims +
		linkshell.BloodsuckerClaims
}

func totalHQClaims(linkshell models.Linkshell) int {
	return linkshell.AspidocheloneClaims +
		linkshell.KingBehemothClaims +
		linkshell.NidhoggClaims
}

func totalGWClaims(linkshell models.Linkshell) int {
	return linkshell.TiamatClaims +
		linkshell.JormungandClaims +
		linkshell.VrtraClaims
}
