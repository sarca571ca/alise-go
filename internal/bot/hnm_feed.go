package bot

// HNM timer feed.
//
// Whenever the HNM Camp Timers board changes, the same list is also written as JSON to a secret
// GitHub Gist, so a game addon (rsvp, for Ashita) can keep its timers up to date without anyone
// copying the board out of Discord.
//
// To turn it on, add one line to .env and restart the bot:
//
//	FEED_GIST_TOKEN=<a GitHub token with the gist scope>
//
// The bot then makes a secret gist for the feed (or finds the one it made before) and posts the
// gist's raw link to the bot log channel each time it starts. Without the token the feed is off
// and the bot works exactly as before.
//
// Optional: FEED_GIST_ID writes to a gist you made yourself instead, and FEED_GIST_FILE names the
// file in the gist (hnm.json when left out).
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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// feedGistDescription marks the gist the bot made, so it finds the same gist after a restart.
const feedGistDescription = "Alise HNM timer feed (read by the rsvp addon)"

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

// gistInfo is the part of GitHub's answer about a gist that the feed uses.
type gistInfo struct {
	ID          string                     `json:"id"`
	Description string                     `json:"description"`
	Files       map[string]json.RawMessage `json:"files"`
	Owner       struct {
		Login string `json:"login"`
	} `json:"owner"`
}

// rawLink is the address that always gives the newest version of the file.
func (g gistInfo) rawLink(file string) string {
	if g.Owner.Login == "" {
		return "https://gist.github.com/" + g.ID
	}

	return "https://gist.githubusercontent.com/" + g.Owner.Login + "/" + g.ID + "/raw/" + file
}

var (
	feedMu     sync.Mutex // one gist write at a time; also guards feedGistID and feedLink
	feedSeq    int64      // counts the lists handed to publishHNMFeed; only the newest is written
	feedGistID string     // the gist in use, found or made on the first write
	feedLink   string     // the raw link last posted to the log channel
	feedHTTP   = &http.Client{Timeout: 20 * time.Second}
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
// updates overlap, only the newer list is written. The feed's link (once per start, or when it
// changes) and any failures go to the bot's log channel.
func (b *Bot) publishHNMFeed(timers []models.HNMTimer) {
	token := strings.TrimSpace(os.Getenv("FEED_GIST_TOKEN"))
	if token == "" {
		return
	}

	file := strings.TrimSpace(os.Getenv("FEED_GIST_FILE"))
	if file == "" {
		file = "hnm.json"
	}

	fixedID := strings.TrimSpace(os.Getenv("FEED_GIST_ID"))

	seq := atomic.AddInt64(&feedSeq, 1)
	feed := buildHNMFeed(timers, time.Now())

	feedMu.Lock()
	defer feedMu.Unlock()

	if seq != atomic.LoadInt64(&feedSeq) {
		return // a newer list has been or will be written
	}

	link, err := syncHNMFeed(token, file, fixedID, feed)
	if err != nil {
		b.logErrorMessage("HNMFeed", err)
		return
	}

	if link != feedLink {
		feedLink = link
		b.logBasicMessage("HNMFeed", "The HNM timer feed is live. Its link for the rsvp addon: "+link)
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

// syncHNMFeed writes the feed and returns the gist's raw link. On the first write it uses
// fixedID when given, else the gist the bot made before, else it makes a new secret gist.
// The caller holds feedMu.
func syncHNMFeed(token, file, fixedID string, feed hnmFeed) (string, error) {
	content, err := json.MarshalIndent(feed, "", "  ")
	if err != nil {
		return "", err
	}

	if fixedID != "" {
		feedGistID = fixedID
	}

	if feedGistID == "" {
		g, found, err := findFeedGist(token, file)
		if err != nil {
			return "", err
		}

		if !found {
			g, err = createFeedGist(token, file, content)
			if err != nil {
				return "", err
			}

			feedGistID = g.ID

			return g.rawLink(file), nil // made with the feed already in it
		}

		feedGistID = g.ID
	}

	g, status, err := writeGistFile(feedGistID, token, file, content)
	if err != nil {
		if status == http.StatusNotFound && fixedID == "" {
			feedGistID = "" // the gist was deleted: find or make one on the next update
		}

		return "", err
	}

	return g.rawLink(file), nil
}

// findFeedGist looks through the token owner's gists for the one the bot made before.
func findFeedGist(token, file string) (gistInfo, bool, error) {
	const perPage = 100

	for page := 1; page <= 30; page++ {
		var list []gistInfo

		path := fmt.Sprintf("/gists?per_page=%d&page=%d", perPage, page)
		if _, err := callGistAPI(http.MethodGet, path, token, nil, &list); err != nil {
			return gistInfo{}, false, fmt.Errorf("looking for the HNM feed gist: %w", err)
		}

		for _, g := range list {
			if _, ok := g.Files[file]; ok && g.Description == feedGistDescription && g.ID != "" {
				return g, true, nil
			}
		}

		if len(list) < perPage {
			break
		}
	}

	return gistInfo{}, false, nil
}

// createFeedGist makes a new secret gist holding the feed.
func createFeedGist(token, file string, content []byte) (gistInfo, error) {
	var g gistInfo

	_, err := callGistAPI(http.MethodPost, "/gists", token, map[string]any{
		"description": feedGistDescription,
		"public":      false,
		"files": map[string]any{
			file: map[string]string{"content": string(content)},
		},
	}, &g)
	if err != nil {
		return g, fmt.Errorf("making the HNM feed gist: %w", err)
	}

	if g.ID == "" {
		return g, errors.New("making the HNM feed gist: GitHub's answer had no gist id")
	}

	return g, nil
}

// writeGistFile replaces one file of a gist. It also returns the HTTP status, so a deleted gist
// (404) can be told apart from other failures.
func writeGistFile(gistID, token, file string, content []byte) (gistInfo, int, error) {
	var g gistInfo

	status, err := callGistAPI(http.MethodPatch, "/gists/"+url.PathEscape(gistID), token, map[string]any{
		"files": map[string]any{
			file: map[string]string{"content": string(content)},
		},
	}, &g)
	if err != nil {
		return g, status, fmt.Errorf("updating the HNM feed gist: %w", err)
	}

	if g.ID == "" {
		g.ID = gistID
	}

	return g, status, nil
}

// callGistAPI sends one request to GitHub's gist API and decodes the answer into out.
func callGistAPI(method, path, token string, body, out any) (int, error) {
	var payload io.Reader

	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, err
		}

		payload = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, gistAPI+path, payload)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "alise-go-hnm-feed")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := feedHTTP.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		msg, _ := io.ReadAll(io.LimitReader(res.Body, 300))
		return res.StatusCode, fmt.Errorf("GitHub answered %s: %s", res.Status, strings.TrimSpace(string(msg)))
	}

	if out != nil {
		if err := json.NewDecoder(io.LimitReader(res.Body, 8<<20)).Decode(out); err != nil {
			return res.StatusCode, fmt.Errorf("reading GitHub's answer: %w", err)
		}
	}

	_, _ = io.Copy(io.Discard, res.Body)

	return res.StatusCode, nil
}
