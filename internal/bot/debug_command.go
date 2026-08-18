package bot

import (
	"alise-go/internal/commands"
	"alise-go/internal/config"
	"alise-go/internal/formatting"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) buildDebugCommand(cfg config.Config) commands.Command {
	debugCmd := commands.DebugCommand{
		OnParsed: func(s *discordgo.Session, i *discordgo.InteractionCreate, window string) {
			_, err := s.Channel(i.ChannelID)
			if err != nil {
				respondEphemeral(s, i, "Failed to load channel.")
				return
			}

			_, _ = b.dg.ChannelMessageSend(i.ChannelID, formatting.FormatWindowHeading(fmt.Sprintf("POP: Window %s | debug", window)))

		},
	}
	return debugCmd
}
