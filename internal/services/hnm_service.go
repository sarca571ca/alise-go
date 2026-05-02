package services

import (
	"alise-go/internal/config"
	"alise-go/internal/data"
	"alise-go/internal/formatting"
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
	go func() {
		for {
			now := time.Now()
			next := now.Truncate(time.Second).Add(time.Second)
			select {
			case <-time.After(time.Until(next)):
				s.tickCamps()
				s.checkTimers()
				s.tickCampWindows()
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
	mins = max(mins, 1)

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
		lastWindow := windows.Windows[len(windows.Windows)-1]

		if !now.Before(lastWindow) {
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

		if _, err = s.store.UpsertHNMCampChannel(chRecord); err != nil {
			log.Println("UpsertHNMCampChannel error:", err)
		}
	}
}

func (s *HNMService) tickCampWindows() {
	guildID := s.cfg.GuildID
	if guildID == "" {
		log.Println("No guildID")
		return
	}

	camps, err := s.store.ListHNMCampChannels(guildID)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	now := time.Now()

	for _, camp := range camps {
		timerChannelID := s.cfg.Channels.HNMTimes
		timerRec, ok, err := s.store.GetHNMTimerRecord(guildID, timerChannelID, camp.HNMID)
		if err != nil {
			log.Println("GetHNMTimerRecord error:", err)
			continue
		}
		if !ok {
			continue
		}

		if shouldArchiveCamp(camp, timerRec) {
			if err := s.store.DeleteHNMCampChannel(camp.ID); err != nil {
				log.Println("DeleteHNMCampChannel error:", err)
			}
			continue
		}

		if camp.IsClosed {
			continue
		}

		hnm, ok := models.GetHNM(camp.HNMID)
		if !ok {
			continue
		}

		timer := data.NewTimerFromRecord(
			data.HNMTimerRecord{
				LastKill:    camp.LastKill,
				DaysSinceHQ: camp.DaysSinceHQ,
			},
			hnm,
		)

		wins := models.BuildHNMTimerWindows(timer)
		if len(wins.Windows) == 0 {
			continue
		}

		lastWin := wins.Windows[len(wins.Windows)-1]

		if camp.MoveScheduled {
			continue
		}

		if now.After(lastWin) && !camp.IsEnraged {
			content := "Moving channel to awaiting-processing in 5 minutes."
			_, _ = s.dg.ChannelMessageSend(camp.ChannelID, content)

			camp.MoveScheduled = true
			if _, err := s.store.UpsertHNMCampChannel(camp); err != nil {
				continue
			}

			go s.MoveCampAfterDelay(camp.ChannelID, 5*time.Minute)
			continue
		}

		idx := currentWindowIndex(now, wins)
		if idx == 0 || idx <= camp.LastWindowIdx {
			continue
		}

		content := formatting.FormatWindowHeading(fmt.Sprintf("Window %d", idx))
		_, err = s.dg.ChannelMessageSend(camp.ChannelID, content)
		if err != nil {
			continue
		}

		camp.LastWindowIdx = idx
		if _, err := s.store.UpsertHNMCampChannel(camp); err != nil {
			continue
		}
	}
}

func (s *HNMService) MoveCampAfterDelay(channelID string, delay time.Duration) {
	time.Sleep(delay)

	guildID := s.cfg.GuildID
	targetParent := s.cfg.Categories.AwaitingProcessingID
	if guildID == "" || targetParent == "" {
		return
	}

	_, _ = s.dg.ChannelEdit(channelID, &discordgo.ChannelEdit{
		ParentID: targetParent,
	})
}
