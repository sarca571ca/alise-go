package bot

import (
	"alise-go/internal/commands"
	"alise-go/internal/config"
	"alise-go/internal/data"
	"alise-go/internal/models"
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

	store, err := data.NewStore("alise.db")
	if err != nil {
		return nil, err
	}
	// store := &data.Store{DB: db}

	reg := commands.NewRegistry()

	b := &Bot{dg: dg, reg: reg, cfg: cfg, store: store}

	hnmCmd := commands.HNMCommand{
		OnParsed: func(s *discordgo.Session, i *discordgo.InteractionCreate, timer models.HNMTimer) {
			guildID := i.GuildID
			channelID := cfg.Channels.HNMTimes

			rec := data.NewRecordFromHNMTimer(guildID, channelID, timer)

			if _, err := b.store.UpsertHNMTimerRecord(rec); err != nil {
				respondError(s, i, "Failed to save timer")
				return
			}

			recs, err := store.ListHNMTimerRecords(guildID, channelID)
			if err != nil {
				respondError(s, i, "Failed to load timers")
			}

			var timers []models.HNMTimer
			for _, r := range recs {
				hnm, ok := models.GetHNM(r.HNMID)
				if !ok {
					continue
				}
				timers = append(timers, data.NewTimerFromRecord(r, hnm))
			}

			if err := b.updateHNMTimerBoard(guildID, timers); err != nil {
				respondError(s, i, "Failed to update timer board")
				return
			}

			content := formatHNMTimersPlain(timers)

			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
				},
			},
			)
		},
	}

	reg.Register(hnmCmd)

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
