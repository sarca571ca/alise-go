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
				s.tickGrandWyrmWindows()
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
		// TODO: need to add a removal of old timers. standard is 3 days old they get removed
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
		if len(windows.Windows) == 0 {
			continue
		}

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
			ch, err := s.dg.Channel(camp.ChannelID)
			if err == nil && ch != nil {
				continue
			}

			if restErr, ok := err.(*discordgo.RESTError); ok &&
				restErr.Response != nil &&
				restErr.Response.StatusCode == 404 {
				camp.ChannelID = ""
				if _, err := s.store.UpsertHNMCampChannel(camp); err != nil {
					log.Println("clear stale ChannelID error:", err)
					continue
				}
			}

			log.Printf("Channel lookup error: campID=%s | channelID=%s | err=%v",
				camp.ID,
				camp.ChannelID,
				err)
			continue
		}

		name, seq, err := s.campNameAndSeq(s.store, guildID, timer)
		if err != nil {
			continue
		}

		reused, err := s.updateLastKill(guildID, timer)
		if err != nil {
			log.Println("updateLastKill error:", err)
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

		if reused && !camp.IsSpawned {
			if _, err = s.updateChannelID(guildID, channel.ID, timer); err != nil {
				log.Println("Record Update Error:", err)
			}
			log.Println("reused the timer and updated the channelid")
		} else if found {
			camp.ChannelID = channel.ID
			camp.Seq = seq
			camp.Mod = timer.Mod

			if _, err = s.store.UpsertHNMCampChannel(camp); err != nil {
				log.Println("Upsert existing camp error:", err)
			}
		} else {
			chRecord := data.HNMCampChannel{
				GuildID:       guildID,
				ChannelID:     channel.ID,
				HNMID:         hnm.ID,
				LastKill:      timer.LastKill,
				DaysSinceHQ:   timer.DaysSinceHQ,
				Seq:           seq,
				Mod:           timer.Mod,
				IsClosed:      false,
				IsEnraged:     false,
				IsSpawned:     false,
				LastWindowIdx: 0,
				MoveScheduled: false,
			}

			if _, err = s.store.UpsertHNMCampChannel(chRecord); err != nil {
				log.Println("UpsertHNMCampChannel error:", err)
			}
		}

		msg := models.GetCampInfo(hnm, timer, firstWindow)

		_, err = s.dg.ChannelMessageSendComplex(channel.ID, msg)
		if err != nil {
			log.Println("EmbedSend Error:", err)
		}

		if !hnm.UseHourlyWarningFlow {
			_, err = s.dg.ChannelMessageSend(
				channel.ID,
				formatting.FormatWindowHeading("Window opens in 20 Minutes x-in"),
			)
			if err != nil {
				log.Println("MsgSend Error: ", err)
			}

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

		hnm, ok := models.GetHNM(camp.HNMID)
		if !ok {
			continue
		}

		if hnm.UseHourlyWarningFlow {
			continue
		}

		if !hnm.KeepCampOpenUntilPop && shouldArchiveCamp(camp, timerRec) {
			if err := s.store.DeleteHNMCampChannel(camp.ID); err != nil {
				log.Println("DeleteHNMCampChannel error:", err)
			}
			continue
		}

		if camp.IsClosed {
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

		idx := currentWindowIndex(now, wins)
		if now.After(lastWin) && !camp.IsEnraged && !hnm.KeepCampOpenUntilPop {
			content := formatting.FormatWindowHeading(fmt.Sprintf("Window %d", camp.LastWindowIdx+1))
			_, _ = s.dg.ChannelMessageSend(camp.ChannelID, content)

			camp.IsClosed = true
			if _, err := s.store.UpsertHNMCampChannel(camp); err != nil {
				continue
			}
			continue
		}

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

func (s *HNMService) tickGrandWyrmWindows() {
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

		hnm, ok := models.GetHNM(camp.HNMID)
		if !ok {
			continue
		}

		if !hnm.UseHourlyWarningFlow {
			continue
		}

		if !hnm.KeepCampOpenUntilPop && shouldArchiveCamp(camp, timerRec) {
			if err := s.store.DeleteHNMCampChannel(camp.ID); err != nil {
				log.Println("DeleteHNMCampChannel error:", err)
			}
			continue
		}

		if camp.IsClosed {
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

		if now.After(lastWin) && !camp.IsEnraged && !hnm.KeepCampOpenUntilPop {
			continue
		}

		respawn := camp.LastKill.Add(hnm.BaseRespawn)

		for i := 1; i <= hnm.WindowCount; i++ {
			windowStart := respawn.Add(time.Duration(i-1) * hnm.WindowInterval)
			warnAt := windowStart.Add(-hnm.WarnBeforeWindow)
			cutoffAt := windowStart.Add(hnm.CutoffAfterWindow)

			if !now.Before(warnAt) && camp.LastWarnedWindowIdx < i {
				_, _ = s.dg.ChannelMessageSend(
					camp.ChannelID,
					formatting.FormatWindowHeading(
						fmt.Sprintf("Window %d opens in 5 minutes x-in", i),
					),
				)
				camp.LastWarnedWindowIdx = i
				if _, err := s.store.UpsertHNMCampChannel(camp); err != nil {
					log.Println("save warning idx error:", err)
				}
			}

			if !now.Before(cutoffAt) && camp.LastCutoffWindowIdx < i {
				_, _ = s.dg.ChannelMessageSend(
					camp.ChannelID,
					formatting.FormatWindowHeading(
						fmt.Sprintf("Window %d closed no more x-in", i),
					),
				)
				camp.LastCutoffWindowIdx = i
				if _, err := s.store.UpsertHNMCampChannel(camp); err != nil {
					log.Println("save cutoff idx error:", err)
				}
			}
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

func (s *HNMService) campNameAndSeq(store *data.Store, guildID string, timer models.HNMTimer) (string, int, error) {
	t := timer.LastKill.Add(timer.HNM.BaseRespawn)

	mon := shortMonth(t)
	day := t.Day()

	short := hnmShortCode(timer.HNM)

	hqSuffix := ""
	if timer.HNM.HQName != "" {
		hqSuffix = fmt.Sprintf("%d", timer.DaysSinceHQ)
	}

	existing, err := store.ListHNMCampChannelsForDay(guildID, timer.HNM.ID, t)
	if err != nil {
		return "", 0, err
	}

	seq := 0
	if len(existing) > 0 {
		for _, c := range existing {
			if _, err := s.dg.Channel(c.ChannelID); err == nil && c.IsSpawned {
				seq = seq + 1
			}
		}
	}

	seqSuffix := ""
	if seq > 0 {
		seqSuffix = fmt.Sprintf("%d", seq)
	}

	name := fmt.Sprintf("%s%d-%s%s%s", mon, day, short, hqSuffix, seqSuffix)
	return name, seq, nil
}

func (s *HNMService) updateChannelID(guildID, channelID string, hnm models.HNMTimer) (bool, error) {
	respawn := hnm.LastKill.Add(hnm.HNM.BaseRespawn)
	existing, err := s.store.ListHNMCampChannelsForDay(guildID, hnm.HNM.ID, respawn)
	if err != nil {
		return false, err
	}

	wins := models.BuildHNMTimerWindows(hnm)
	now := time.Now()
	for _, c := range existing {
		if _, err := s.dg.Channel(c.ChannelID); err != nil && !c.IsSpawned {
			c.ChannelID = channelID
			c.LastWindowIdx = lastCompletedWindowIndex(now, wins)
			if _, err = s.store.UpsertHNMCampChannel(c); err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return false, nil
}

func (s *HNMService) updateLastKill(guildID string, hnm models.HNMTimer) (bool, error) {
	existing, err := s.store.ListHNMCampChannelsForDay(guildID, hnm.HNM.ID, hnm.LastKill.Add(hnm.HNM.BaseRespawn))
	if err != nil {
		return false, err
	}

	for _, c := range existing {
		rec, ok, err := s.store.GetHNMCampChannelByChannelID(guildID, c.ChannelID)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, fmt.Errorf("record not found")
		}

		rec.LastKill = hnm.LastKill
		rec.DaysSinceHQ = hnm.DaysSinceHQ

		if _, err = s.store.UpsertHNMCampChannel(rec); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func lastCompletedWindowIndex(now time.Time, wins models.HNMTimerWindows) int {
	idx := currentWindowIndex(now, wins)
	if idx <= 0 {
		return 0
	}
	return idx - 1
}
