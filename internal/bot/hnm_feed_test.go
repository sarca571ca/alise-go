package bot

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

type gistRequest struct {
	method, path, auth, accept, version string
	files                               map[string]string
}

// gistServer answers like GitHub's "update a gist" call and hands each request it gets to the test.
func gistServer(t *testing.T, status int, reply string) (<-chan gistRequest, func()) {
	t.Helper()

	got := make(chan gistRequest, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Files map[string]struct {
				Content string `json:"content"`
			} `json:"files"`
		}

		req := gistRequest{
			method:  r.Method,
			path:    r.URL.Path,
			auth:    r.Header.Get("Authorization"),
			accept:  r.Header.Get("Accept"),
			version: r.Header.Get("X-GitHub-Api-Version"),
			files:   map[string]string{},
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			for name, f := range body.Files {
				req.files[name] = f.Content
			}
		}

		got <- req

		w.WriteHeader(status)
		_, _ = io.WriteString(w, reply)
	}))

	old := gistAPI
	gistAPI = srv.URL

	return got, func() {
		gistAPI = old
		srv.Close()
	}
}

func TestWriteGistFile(t *testing.T) {
	got, done := gistServer(t, http.StatusOK, `{}`)
	defer done()

	feed := buildHNMFeed([]models.HNMTimer{
		feedTimer(t, "adamantoise", time.Unix(1795483233, 0), 8, ""),
	}, time.Unix(1795535090, 0))

	if err := writeGistFile("abc123", "secret-token", "hnm.json", feed); err != nil {
		t.Fatal(err)
	}

	req := <-got

	if req.method != http.MethodPatch || req.path != "/gists/abc123" {
		t.Errorf("request: got %s %s, want PATCH /gists/abc123", req.method, req.path)
	}

	if req.auth != "Bearer secret-token" || req.accept != "application/vnd.github+json" || req.version != "2022-11-28" {
		t.Errorf("headers: Authorization %q, Accept %q, X-GitHub-Api-Version %q", req.auth, req.accept, req.version)
	}

	content, ok := req.files["hnm.json"]
	if !ok || len(req.files) != 1 {
		t.Fatalf("files sent: %v, want only hnm.json", req.files)
	}

	var back hnmFeed
	if err := json.Unmarshal([]byte(content), &back); err != nil {
		t.Fatalf("hnm.json is not valid JSON: %v\n%s", err, content)
	}

	if back.Updated != feed.Updated || len(back.HNMs) != 1 || back.HNMs[0] != feed.HNMs[0] {
		t.Errorf("hnm.json reads back as %+v, want %+v", back, feed)
	}

	if back.HNMs[0].Name != "Adamantoise/Aspidochelone" || back.HNMs[0].Next != 1795483233+22*3600 {
		t.Errorf("Adamantoise on day 8: got %+v", back.HNMs[0])
	}
}

func TestWriteGistFileReportsGitHubErrors(t *testing.T) {
	got, done := gistServer(t, http.StatusUnauthorized, `{"message":"Bad credentials"}`)
	defer done()

	err := writeGistFile("abc123", "expired", "hnm.json", buildHNMFeed(nil, time.Now()))
	<-got

	if err == nil || !strings.Contains(err.Error(), "401") || !strings.Contains(err.Error(), "Bad credentials") {
		t.Errorf("got error %v, want one naming the 401 and GitHub's message", err)
	}
}
