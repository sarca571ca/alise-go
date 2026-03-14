package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) registerHandlers() {
	b.dg.AddHandler(b.onReady)
	b.dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		b.reg.HandleInteraction(s, i)
	})
}

func (b *Bot) onReady(s *discordgo.Session, r *discordgo.Ready) {
	fmt.Println("Bot is running. Press Ctrl+C to exit.")
}
