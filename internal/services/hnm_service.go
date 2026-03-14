package services

import (
	"alise-go/internal/config"
	"alise-go/internal/data"
	"alise-go/internal/models"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

type HNMService struct {
	store *data.Store
	cfg   config.Config
	dg    *discordgo.Session
}

func NewHNMService(store *data.Store, cfg config.Config, dg *discordgo.Session) *HNMService {
	return &HNMService{
		store: store,
		cfg:   cfg,
		dg:    dg,
	}
}

func (s *HNMService) StartPolling(stop <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.tickCamps()
				s.checkTimers()
			case <-stop:
				return
			}
		}
	}()
}

func (s *HNMService) checkTimers() {
	guildID := s.cfg.GuildID
	channelID := s.cfg.Channels.HNMTimes

	recs, err := s.store.ListHNMTimerRecords(guildID, channelID)
	if err != nil {
		return
	}

	now := time.Now()
	oneHour := time.Hour

	for _, r := range recs {
		hnm, ok := models.GetHNM(r.HNMID)
		if !ok {
			continue
		}

		if r.IsNotified {
			continue
		}

		timer := data.NewTimerFromRecord(r, hnm)
		windows := models.BuildHNMTimerWindows(timer)

		next := windows.NextRespawn
		diff := next.Sub(now)
		if diff <= 0 || diff > oneHour {
			continue
		}

		s.sendCampPing(hnm.Name, next, diff)

		r.IsNotified = true
		if _, err := s.store.UpsertHNMTimerRecord(r); err != nil {
			continue
		}
	}
}

func (s *HNMService) sendCampPing(hnmName string, spawn time.Time, diff time.Duration) {
	channelID := s.cfg.Channels.CampPings
	if channelID == "" {
		return
	}

	mins := int(diff.Minutes())
	if mins < 1 {
		mins = 1
	}

	content := fmt.Sprintf(
		"@everyone %s camp will start in %d minutes (respawn at <t:%d:R>)",
		hnmName,
		mins,
		spawn.Unix(),
	)

	_, _ = s.dg.ChannelMessageSend(channelID, content)
}

func (s *HNMService) tickCamps() {
	guildID := s.cfg.GuildID
	timerChannelID := s.cfg.Channels.HNMTimes
	if guildID == "" || timerChannelID == "" {
		return
	}

	recs, err := s.store.ListHNMTimerRecords(guildID, timerChannelID)
	if err != nil {
		return
	}

	now := time.Now()

	for _, r := range recs {
		hnm, ok := models.GetHNM(r.HNMID)
		if !ok {
			continue
		}
		timer := data.NewTimerFromRecord(r, hnm)
		windows := models.BuildHNMTimerWindows(timer)

		firstWindow := windows.NextRespawn

		if !now.Before(firstWindow) {
			continue
		}

		if now.Before(firstWindow.Add(-20 * time.Minute)) {
			continue
		}

		camp, found, err := s.store.GetHNMCampChannelByCamp(
			guildID,
			hnm.ID,
			timer.LastKill,
			timer.DaysSinceHQ,
		)
		if err != nil {
			continue
		}
		if found && camp.ChannelID != "" {
			// Camp channel already exists for this kill.
			continue
		}

		name, seq, err := campNameAndSeq(s.store, guildID, timer)
		if err != nil {
			continue
		}

		parentID := s.cfg.Categories.HNMCategoryID

		channel, err := s.dg.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
			Name:     name,
			Type:     discordgo.ChannelTypeGuildText,
			ParentID: parentID,
		})
		if err != nil {
			continue
		}

		content := fmt.Sprintf(
			"Camp channel created for %s (HQ day %d). First window at <t:%d:R>.",
			hnm.Name,
			timer.DaysSinceHQ,
			firstWindow.Unix(),
		)
		_, _ = s.dg.ChannelMessageSend(channel.ID, content)

		chRecord := data.HNMCampChannel{
			GuildID:       guildID,
			ChannelID:     channel.ID,
			HNMID:         hnm.ID,
			LastKill:      timer.LastKill,
			DaysSinceHQ:   timer.DaysSinceHQ,
			Seq:           seq,
			IsClosed:      false,
			IsEnraged:     false,
			LastWindowIdx: 0,
		}
		log.Printf("existingToday for %s: found=%+v, channelID=%q\n", hnm.ID, found, camp.ChannelID)

		_, _ = s.store.UpsertHNMCampChannel(chRecord)
	}
}
