package bot

import (
	"alise-go/internal/commands"
	"alise-go/internal/config"
	"alise-go/internal/data"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	dg    *discordgo.Session
	reg   *commands.Registery
	cfg   config.Config
	store *data.Store
}

func New(cfg config.Config) (*Bot, error) {
	dg, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return nil, err
	}

	reg := commands.NewRegistry()
	reg.Register(commands.HNMCommand{})

	dg.Identify.Intents = discordgo.IntentGuilds

	b := &Bot{dg: dg, reg: reg, cfg: cfg}
	return b, nil
}

func (b *Bot) registerSlashCommands(appID string) error {
	defs := b.reg.SlashDefs()

	scopeGuildID := b.cfg.GuildID

	log.Println("Bulk overwritting slash commands for guild:", scopeGuildID)

	_, err := b.dg.ApplicationCommandBulkOverwrite(appID, scopeGuildID, defs)
	if err != nil {
		log.Println("error registering slash commands:", err)
	}

	return err
}

func (b *Bot) Start(ctx context.Context) error {
	b.registerHandlers()
	if err := b.dg.Open(); err != nil {
		return err
	}
	defer b.dg.Close()

	appID := b.dg.State.User.ID

	if err := b.registerSlashCommands(appID); err != nil {
		return err
	}

	<-waitForSignal(ctx)

	return nil
}

func waitForSignal(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	go func() {
		defer close(done)
		select {
		case <-ctx.Done():
		case <-ch:
		}
	}()

	return done
}
