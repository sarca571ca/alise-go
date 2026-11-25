package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"alise-go/internal/models"
)

func feedTimer(t *testing.T, id string, kill time.Time, day int, mod string) models.HNMTimer {
	t.Helper()

	hnm, ok := models.GetHNM(id)
	if !ok {
		t.Fatalf("unknown HNM %q", id)
	}

	return models.HNMTimer{HNM: hnm, Mod: mod, LastKill: kill, DaysSinceHQ: day}
}

func TestBuildHNMFeed(t *testing.T) {
	kill := time.Unix(1795464039, 0)
	now := time.Unix(1795535090, 0)

	feed := buildHNMFeed([]models.HNMTimer{
		feedTimer(t, "tiamat", kill, 0, "n"),
		feedTimer(t, "behemoth", kill, 4, "d"),
		feedTimer(t, "fafnir", kill.Add(time.Hour), 3, ""),
		feedTimer(t, "shiki", kill, 0, "a"),
	}, now)

	if feed.Updated != now.Unix() {
		t.Errorf("updated: got %d, want %d", feed.Updated, now.Unix())
	}

	// Sorted by next window; the board's names and days; mods other than normal kept.
	want := []hnmFeedEntry{
		{Name: "Shikigami Weapon", Next: kill.Add(21 * time.Hour).Unix(), ToD: kill.Unix(), Windows: 1, Interval: 0, Mod: "a"},
		{Name: "Behemoth/King Behemoth", Next: kill.Add(22 * time.Hour).Unix(), ToD: kill.Unix(), Day: 4, Windows: 7, Interval: 600, Mod: "d"},
		{Name: "Fafnir", Next: kill.Add(23 * time.Hour).Unix(), ToD: kill.Add(time.Hour).Unix(), Day: 3, Windows: 7, Interval: 600},
		{Name: "Tiamat", Next: kill.Add(84 * time.Hour).Unix(), ToD: kill.Unix(), Windows: 25, Interval: 3600},
	}

	if len(feed.HNMs) != len(want) {
		t.Fatalf("got %d entries, want %d", len(feed.HNMs), len(want))
	}

	for i := range want {
		if feed.HNMs[i] != want[i] {
			t.Errorf("entry %d:\n got  %+v\n want %+v", i, feed.HNMs[i], want[i])
		}
	}
}

func TestBuildHNMFeedEmpty(t *testing.T) {
	out, err := json.Marshal(buildHNMFeed(nil, time.Unix(1795535090, 0)))
	if err != nil {
		t.Fatal(err)
	}

	if got := string(out); got != `{"updated":1795535090,"hnms":[]}` {
		t.Errorf("an empty board gives %s, want an empty list", got)
	}
}

// fakeGitHub answers like GitHub's gist API (list, create, update), keeping the gists in memory.
type fakeGitHub struct {
	t        *testing.T
	mu       sync.Mutex
	gists    []*fakeGist // newest first, as GitHub lists them
	made     int
	requests []seenRequest
	fail     int // when not 0, every request gets this status and failBody
	failBody string
}

type fakeGist struct {
	id, description, owner string
	public                 bool
	files                  map[string]string
}

type seenRequest struct {
	call                  string // e.g. "PATCH /gists/abc123"
	auth, accept, version string
	description           *string
	public                *bool
	files                 map[string]string
}

func newFakeGitHub(t *testing.T) *fakeGitHub {
	t.Helper()

	f := &fakeGitHub{t: t}
	srv := httptest.NewServer(f)

	old := gistAPI
	gistAPI = srv.URL
	resetFeedState()

	t.Cleanup(func() {
		gistAPI = old
		srv.Close()
		resetFeedState()
	})

	return f
}

func resetFeedState() {
	feedMu.Lock()
	defer feedMu.Unlock()

	feedGistID = ""
	feedLink = ""
}

func (g *fakeGist) reply(withContent bool) map[string]any {
	files := map[string]any{}
	for name, content := range g.files {
		f := map[string]any{"filename": name, "raw_url": "https://gist.githubusercontent.com/" + g.owner + "/" + g.id + "/raw/0123abcd/" + name}
		if withContent {
			f["content"] = content
		}
		files[name] = f
	}

	return map[string]any{
		"id":          g.id,
		"description": g.description,
		"public":      g.public,
		"owner":       map[string]any{"login": g.owner},
		"files":       files,
	}
}

func (f *fakeGitHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	var body struct {
		Description *string `json:"description"`
		Public      *bool   `json:"public"`
		Files       map[string]struct {
			Content string `json:"content"`
		} `json:"files"`
	}

	if raw, _ := io.ReadAll(r.Body); len(raw) > 0 {
		if err := json.Unmarshal(raw, &body); err != nil {
			f.t.Errorf("%s %s: the body is not JSON: %v", r.Method, r.URL.Path, err)
		}
	}

	call := r.Method + " " + r.URL.Path
	if r.URL.RawQuery != "" {
		call += "?" + r.URL.RawQuery
	}

	seen := seenRequest{
		call:        call,
		auth:        r.Header.Get("Authorization"),
		accept:      r.Header.Get("Accept"),
		version:     r.Header.Get("X-GitHub-Api-Version"),
		description: body.Description,
		public:      body.Public,
		files:       map[string]string{},
	}
	for name, file := range body.Files {
		seen.files[name] = file.Content
	}
	f.requests = append(f.requests, seen)

	w.Header().Set("Content-Type", "application/json")

	if f.fail != 0 {
		w.WriteHeader(f.fail)
		_, _ = io.WriteString(w, f.failBody)
		return
	}

	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/gists":
		perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
		if perPage == 0 {
			perPage = 30
		}
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page == 0 {
			page = 1
		}

		list := []map[string]any{}
		for i := (page - 1) * perPage; i < page*perPage && i < len(f.gists); i++ {
			list = append(list, f.gists[i].reply(false))
		}
		_ = json.NewEncoder(w).Encode(list)

	case r.Method == http.MethodPost && r.URL.Path == "/gists":
		f.made++
		g := &fakeGist{id: fmt.Sprintf("made%d", f.made), owner: "aliseowner", files: map[string]string{}}
		if body.Description != nil {
			g.description = *body.Description
		}
		if body.Public != nil {
			g.public = *body.Public
		}
		for name, file := range body.Files {
			g.files[name] = file.Content
		}
		f.gists = append([]*fakeGist{g}, f.gists...)

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(g.reply(true))

	case r.Method == http.MethodPatch && strings.HasPrefix(r.URL.Path, "/gists/"):
		g := f.find(strings.TrimPrefix(r.URL.Path, "/gists/"))
		if g == nil {
			w.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(w, `{"message":"Not Found"}`)
			return
		}
		for name, file := range body.Files {
			g.files[name] = file.Content
		}
		_ = json.NewEncoder(w).Encode(g.reply(true))

	default:
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"message":"Not Found"}`)
	}
}

func (f *fakeGitHub) find(id string) *fakeGist {
	for _, g := range f.gists {
		if g.id == id {
			return g
		}
	}
	return nil
}

func (f *fakeGitHub) add(g *fakeGist) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.gists = append(f.gists, g)
}

func (f *fakeGitHub) remove(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i, g := range f.gists {
		if g.id == id {
			f.gists = append(f.gists[:i], f.gists[i+1:]...)
			return
		}
	}
}

// takeCalls returns the calls made since the last takeCalls, checking every one's headers.
func (f *fakeGitHub) takeCalls(token string) ([]string, []seenRequest) {
	f.mu.Lock()
	defer f.mu.Unlock()

	var calls []string
	for _, r := range f.requests {
		calls = append(calls, r.call)
		if r.auth != "Bearer "+token || r.accept != "application/vnd.github+json" || r.version != "2022-11-28" {
			f.t.Errorf("%s: headers Authorization %q, Accept %q, X-GitHub-Api-Version %q", r.call, r.auth, r.accept, r.version)
		}
	}

	reqs := f.requests
	f.requests = nil

	return calls, reqs
}

func wantCalls(t *testing.T, step string, got []string, want ...string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s: calls\n got  %q\n want %q", step, got, want)
	}
}

func readFeed(t *testing.T, content string) hnmFeed {
	t.Helper()

	var feed hnmFeed
	if err := json.Unmarshal([]byte(content), &feed); err != nil {
		t.Fatalf("hnm.json is not valid JSON: %v\n%s", err, content)
	}

	return feed
}

func TestFeedMakesItsOwnSecretGistThenUpdatesIt(t *testing.T) {
	gh := newFakeGitHub(t)

	// Another gist with the same file name, but not the bot's: it must be left alone.
	gh.add(&fakeGist{id: "notes1", description: "my notes", owner: "aliseowner", files: map[string]string{"hnm.json": "{}"}})

	first := buildHNMFeed([]models.HNMTimer{
		feedTimer(t, "adamantoise", time.Unix(1795483233, 0), 8, ""),
	}, time.Unix(1795535090, 0))

	link, err := syncHNMFeed("secret-token", "hnm.json", "", first)
	if err != nil {
		t.Fatal(err)
	}

	calls, reqs := gh.takeCalls("secret-token")
	wantCalls(t, "first write", calls, "GET /gists?per_page=100&page=1", "POST /gists")

	if link != "https://gist.githubusercontent.com/aliseowner/made1/raw/hnm.json" {
		t.Errorf("link: got %s", link)
	}

	if len(reqs) == 2 {
		made := reqs[1]
		if made.public == nil || *made.public || made.description == nil || *made.description != feedGistDescription {
			t.Errorf("the new gist must be secret and carry the feed's description: public %v, description %v", made.public, made.description)
		}

		back := readFeed(t, made.files["hnm.json"])
		if len(made.files) != 1 || back.Updated != first.Updated || len(back.HNMs) != 1 || back.HNMs[0] != first.HNMs[0] {
			t.Errorf("the new gist holds %v, want only hnm.json with %+v", made.files, first)
		}

		if back.HNMs[0].Name != "Adamantoise/Aspidochelone" || back.HNMs[0].Next != 1795483233+22*3600 {
			t.Errorf("Adamantoise on day 8: got %+v", back.HNMs[0])
		}
	}

	// The next update goes straight to the same gist.
	second := buildHNMFeed([]models.HNMTimer{
		feedTimer(t, "tiamat", time.Unix(1795486979, 0), 0, "d"),
	}, time.Unix(1795536000, 0))

	link2, err := syncHNMFeed("secret-token", "hnm.json", "", second)
	if err != nil {
		t.Fatal(err)
	}

	calls, reqs = gh.takeCalls("secret-token")
	wantCalls(t, "second write", calls, "PATCH /gists/made1")

	if link2 != link {
		t.Errorf("the link changed from %s to %s", link, link2)
	}

	if len(reqs) == 1 {
		back := readFeed(t, reqs[0].files["hnm.json"])
		if back.Updated != second.Updated || len(back.HNMs) != 1 || back.HNMs[0] != second.HNMs[0] {
			t.Errorf("the update sent %+v, want %+v", back, second)
		}
	}

	if g := gh.find("notes1"); g.files["hnm.json"] != "{}" {
		t.Errorf("the other gist was changed: %q", g.files["hnm.json"])
	}
}

func TestFeedFindsItsGistAfterARestart(t *testing.T) {
	gh := newFakeGitHub(t)

	// 150 gists, so the bot's is on the second page of the list.
	for i := 0; i < 150; i++ {
		g := &fakeGist{id: fmt.Sprintf("other%d", i), owner: "aliseowner", files: map[string]string{"notes.txt": ""}}
		if i == 120 {
			g = &fakeGist{id: "feed42", description: feedGistDescription, owner: "aliseowner", files: map[string]string{"hnm.json": "{}"}}
		}
		gh.add(g)
	}

	link, err := syncHNMFeed("tok", "hnm.json", "", buildHNMFeed(nil, time.Unix(1795535090, 0)))
	if err != nil {
		t.Fatal(err)
	}

	calls, _ := gh.takeCalls("tok")
	wantCalls(t, "after a restart", calls, "GET /gists?per_page=100&page=1", "GET /gists?per_page=100&page=2", "PATCH /gists/feed42")

	if link != "https://gist.githubusercontent.com/aliseowner/feed42/raw/hnm.json" {
		t.Errorf("link: got %s", link)
	}

	if got := gh.find("feed42").files["hnm.json"]; got != "{\n  \"updated\": 1795535090,\n  \"hnms\": []\n}" {
		t.Errorf("the gist now holds %q", got)
	}
}

func TestFeedUsesFeedGistIDWhenGiven(t *testing.T) {
	gh := newFakeGitHub(t)
	gh.add(&fakeGist{id: "abc123", owner: "someone", files: map[string]string{"timers.json": ""}})

	link, err := syncHNMFeed("tok", "timers.json", "abc123", buildHNMFeed(nil, time.Unix(1795535090, 0)))
	if err != nil {
		t.Fatal(err)
	}

	calls, _ := gh.takeCalls("tok")
	wantCalls(t, "with FEED_GIST_ID", calls, "PATCH /gists/abc123")

	if link != "https://gist.githubusercontent.com/someone/abc123/raw/timers.json" {
		t.Errorf("link: got %s", link)
	}
}

func TestFeedMakesANewGistWhenItsGistWasDeleted(t *testing.T) {
	gh := newFakeGitHub(t)
	feed := buildHNMFeed(nil, time.Unix(1795535090, 0))

	if _, err := syncHNMFeed("tok", "hnm.json", "", feed); err != nil {
		t.Fatal(err)
	}
	gh.takeCalls("tok")

	gh.remove("made1")

	if _, err := syncHNMFeed("tok", "hnm.json", "", feed); err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("writing to the deleted gist: got error %v, want a 404", err)
	}

	link, err := syncHNMFeed("tok", "hnm.json", "", feed)
	if err != nil {
		t.Fatal(err)
	}

	calls, _ := gh.takeCalls("tok")
	wantCalls(t, "after the gist was deleted", calls, "PATCH /gists/made1", "GET /gists?per_page=100&page=1", "POST /gists")

	if link != "https://gist.githubusercontent.com/aliseowner/made2/raw/hnm.json" {
		t.Errorf("link: got %s", link)
	}
}

func TestFeedReportsGitHubErrors(t *testing.T) {
	gh := newFakeGitHub(t)
	gh.fail = http.StatusUnauthorized
	gh.failBody = `{"message":"Bad credentials"}`

	_, err := syncHNMFeed("expired", "hnm.json", "", buildHNMFeed(nil, time.Now()))
	if err == nil || !strings.Contains(err.Error(), "401") || !strings.Contains(err.Error(), "Bad credentials") {
		t.Errorf("got error %v, want one naming the 401 and GitHub's message", err)
	}

	_, err = syncHNMFeed("expired", "hnm.json", "abc123", buildHNMFeed(nil, time.Now()))
	if err == nil || !strings.HasPrefix(err.Error(), "updating the HNM feed gist: GitHub answered 401") {
		t.Errorf("with FEED_GIST_ID: got error %v", err)
	}
}
