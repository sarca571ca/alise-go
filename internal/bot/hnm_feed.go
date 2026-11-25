package bot

// HNM timer feed.
//
// Whenever the HNM Camp Timers board changes, the same list is also written as JSON to a GitHub
// Gist, so a game addon (rsvp, for Ashita) can keep its timers up to date without anyone copying
// the board out of Discord.
//
// To turn it on, add two lines to .env and restart the bot:
//
//	FEED_GIST_ID=<the gist's id: the last part of its address>
//	FEED_GIST_TOKEN=<a token that can write gists>
//
// FEED_GIST_FILE names the file in the gist (hnm.json when left out). With FEED_GIST_ID or
// FEED_GIST_TOKEN missing, the feed is off and the bot works exactly as before.
//
// The feed (times in unix seconds, sorted like the board; "day" only for the HNMs that can HQ,
// "mod" only when the camp has one: d double DKP, t triple, a approximate):
//
//	{
//	  "updated": 1795535090,
//	  "hnms": [
//	    { "name": "Behemoth/King Behemoth", "next": 1795543239, "tod": 1795464039,
//	      "day": 4, "windows": 7, "interval": 600 },
//	    ...
//	  ]
//	}

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"alise-go/internal/data"
	"alise-go/internal/models"
)

// gistAPI is GitHub's API address. A variable so the tests can use a local server instead.
var gistAPI = "https://api.github.com"

type hnmFeedEntry struct {
	Name     string `json:"name"`          // as the board shows it, e.g. "Behemoth/King Behemoth"
	Next     int64  `json:"next"`          // first window
	ToD      int64  `json:"tod"`           // last kill
	Day      int    `json:"day,omitempty"` // the board's day number (Fafnir, Adamantoise, Behemoth)
	Windows  int    `json:"windows"`       // how many windows
	Interval int64  `json:"interval"`      // seconds between windows (0 with one window)
	Mod      string `json:"mod,omitempty"` // d, t or a; left out for a normal camp
}

type hnmFeed struct {
	Updated int64          `json:"updated"`
	HNMs    []hnmFeedEntry `json:"hnms"`
}

var (
	feedMu   sync.Mutex // one gist write at a time
	feedSeq  int64      // counts the lists handed to publishHNMFeed; only the newest is written
	feedHTTP = &http.Client{Timeout: 20 * time.Second}
)

// buildHNMFeed turns the board's timers into the feed, with the names, days and order the board uses.
func buildHNMFeed(timers []models.HNMTimer, now time.Time) hnmFeed {
	feed := hnmFeed{
		Updated: now.Unix(),
		HNMs:    make([]hnmFeedEntry, 0, len(timers)),
	}

	for _, t := range timers {
		w := models.BuildHNMTimerWindows(t)

		e := hnmFeedEntry{
			Name:     t.HNM.Name,
			Next:     w.NextRespawn.Unix(),
			ToD:      t.LastKill.Unix(),
			Windows:  t.HNM.WindowCount,
			Interval: int64(t.HNM.WindowInterval / time.Second),
		}

		// Same rule as BuildHNMTimerName: the day for HNMs that can HQ, and both names after day 3.
		if models.CanHQ(t.HNM.Name) {
			e.Day = t.DaysSinceHQ

			if t.DaysSinceHQ > 3 && t.HNM.HQName != "" {
				e.Name = t.HNM.Name + "/" + t.HNM.HQName
			}
		}

		if mod := strings.ToLower(strings.TrimSpace(t.Mod)); mod != "" && mod != "n" {
			e.Mod = mod
		}

		feed.HNMs = append(feed.HNMs, e)
	}

	sort.SliceStable(feed.HNMs, func(i, j int) bool {
		return feed.HNMs[i].Next < feed.HNMs[j].Next
	})

	return feed
}

// publishHNMFeed writes the timers to the gist. Call it as `go b.publishHNMFeed(timers)`, so a slow
// GitHub never holds up the bot's reply to Discord (which has to come within 3 seconds). When two
// updates overlap, only the newer list is written. Failures go to the bot's log channel.
func (b *Bot) publishHNMFeed(timers []models.HNMTimer) {
	gistID := strings.TrimSpace(os.Getenv("FEED_GIST_ID"))
	token := strings.TrimSpace(os.Getenv("FEED_GIST_TOKEN"))

	if gistID == "" || token == "" {
		return
	}

	file := strings.TrimSpace(os.Getenv("FEED_GIST_FILE"))
	if file == "" {
		file = "hnm.json"
	}

	seq := atomic.AddInt64(&feedSeq, 1)
	feed := buildHNMFeed(timers, time.Now())

	feedMu.Lock()
	defer feedMu.Unlock()

	if seq != atomic.LoadInt64(&feedSeq) {
		return // a newer list has been or will be written
	}

	if err := writeGistFile(gistID, token, file, feed); err != nil {
		b.logErrorMessage("HNMFeed", err)
	}
}

// refreshHNMFeed publishes the timers saved for the HNM times channel. Call it once at start-up
// (`go b.refreshHNMFeed()`), so the feed is filled straight away and catches up after downtime.
func (b *Bot) refreshHNMFeed() {
	guildID := b.cfg.GuildID
	channelID := b.cfg.Channels.HNMTimes

	// Without both, this can't tell which timers are the board's; wait for the next board update.
	if guildID == "" || channelID == "" {
		return
	}

	recs, err := b.store.ListHNMTimerRecords(guildID, channelID)
	if err != nil {
		b.logErrorMessage("HNMFeed", err)
		return
	}

	var timers []models.HNMTimer
	for _, r := range recs {
		hnm, ok := models.GetHNM(r.HNMID)
		if !ok {
			continue
		}
		timers = append(timers, data.NewTimerFromRecord(r, hnm))
	}

	b.publishHNMFeed(timers)
}

// writeGistFile replaces one file of a gist with the feed.
func writeGistFile(gistID, token, file string, feed hnmFeed) error {
	content, err := json.MarshalIndent(feed, "", "  ")
	if err != nil {
		return err
	}

	body, err := json.Marshal(map[string]any{
		"files": map[string]any{
			file: map[string]string{"content": string(content)},
		},
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPatch, gistAPI+"/gists/"+gistID, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "alise-go-hnm-feed")

	res, err := feedHTTP.Do(req)
	if err != nil {
		return fmt.Errorf("updating the HNM feed gist: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(res.Body, 300))
		return fmt.Errorf("updating the HNM feed gist: GitHub answered %s: %s", res.Status, strings.TrimSpace(string(msg)))
	}

	_, _ = io.Copy(io.Discard, res.Body)

	return nil
}
