package bot

import (
	"alise-go/internal/models"
	"fmt"
	"strings"
)

func formatLinkshellListPlain(linkshells []models.Linkshell) string {
	// TODO: This will need more expansion/formating fixes but its a working concept for now.
	if len(linkshells) == 0 {
		return "No linkshells currently registered."
	}

	var sb strings.Builder
	sb.WriteString("Claims Leaderboards\n")
	sb.WriteString("Linkshell\tFafnir(**Nidhogg**)\tAdamantoise(**Aspidochelone**)\tBehemoth(**King Behemoth**)\n")

	for _, ls := range linkshells {
		fmt.Fprintf(&sb,
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

	return sb.String()
}
