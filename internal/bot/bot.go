package bot

import (
	"alise-go/internal/commands"
	"alise-go/internal/config"
	"alise-go/internal/data"
	"alise-go/internal/services"
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
	hnm   *services.HNMService
}

func New(cfg config.Config) (*Bot, error) {
	dg, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return nil, err
	}

	store, err := data.NewStore("alise.db")
	if err != nil {
		return nil, err
	}

	reg := commands.NewRegistry()

	b := &Bot{dg: dg, reg: reg, cfg: cfg, store: store}

	cmds := b.buildCommands(cfg)

	for _, cmd := range cmds {
		reg.Register(cmd)
	}

	b.hnm = services.NewHNMService(store, cfg, dg)

	dg.Identify.Intents = discordgo.IntentGuilds

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

	stop := make(chan struct{})
	b.hnm.StartPolling(stop)

	<-waitForSignal(ctx)
	close(stop)

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
