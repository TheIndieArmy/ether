package arr

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/braedonsaunders/ether/internal/config"
	"github.com/braedonsaunders/ether/internal/core"
)

func TestEpisodeFileByShow(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v3/series", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("tvdbId") != "42" {
			t.Fatalf("tvdbId %q", r.URL.Query().Get("tvdbId"))
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 7, "title": "Show", "tvdbId": 42, "tmdbId": 100}})
	})
	mux.HandleFunc("/api/v3/episode", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("seriesId") != "7" || r.URL.Query().Get("seasonNumber") != "1" {
			t.Fatalf("query %v", r.URL.Query())
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"id": 99, "seasonNumber": 1, "episodeNumber": 2, "hasFile": true, "title": "Pilot",
			"episodeFile": map[string]any{"path": "/tv/Show/S01E02.mkv"},
		}})
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	c := New(ts.URL, "key", "sonarr")
	mf, err := c.EpisodeFileByShow(t.Context(), 42, 0, 1, 2)
	if err != nil || mf == nil || mf.ArrPath == "" || mf.EpisodeID != 99 || mf.Kind != "sonarr" {
		t.Fatalf("episode: %+v %v", mf, err)
	}
}

func TestLookupAndDiscoverLocal(t *testing.T) {
	dir := t.TempDir()
	arrPath := filepath.Join(dir, "arr", "Fight Club.mkv")
	localPath := filepath.Join(dir, "local", "Fight Club.mkv")
	if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(localPath, []byte("video-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v3/movie", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"id": 9, "title": "Fight Club", "hasFile": true,
			"movieFile": map[string]any{"path": arrPath},
		}})
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	cfg := config.Arr{
		Enabled:  true,
		Radarr:   config.ArrInstance{URL: ts.URL, APIKey: "k"},
		PathMaps: []config.ArrPathMap{{ArrPrefix: filepath.Join(dir, "arr"), LocalPrefix: filepath.Join(dir, "local")}},
	}
	title := &core.Title{Type: core.TypeMovie, Key: core.MovieKey(550), IDs: core.IDs{TMDB: 550}, Title: "Fight Club"}
	mf, err := Lookup(t.Context(), title, cfg)
	if err != nil || mf == nil || !mf.HasFile {
		t.Fatalf("lookup: %+v %v", mf, err)
	}
	rs := Discover(t.Context(), title, cfg, DiscoverOptions{})
	if len(rs) != 1 || rs[0].Protocol != core.ProtocolLocal || rs[0].StreamURL != localPath {
		t.Fatalf("discover: %+v", rs)
	}

	payload, err := PayloadFromRelease(rs[0])
	if err != nil || !IsLocalPayload(payload) {
		t.Fatalf("payload: %v", err)
	}
	eng := WrapEngine(stubEngine{})
	src, err := eng.Open(t.Context(), rs[0].ID, payload)
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	if src.Size() != int64(len("video-bytes")) {
		t.Fatalf("size %d", src.Size())
	}
	ok, err := eng.CheckRelease(t.Context(), rs[0].ID, payload)
	if err != nil || !ok {
		t.Fatalf("health: %v %v", ok, err)
	}
}

type stubEngine struct{}

func (stubEngine) Open(context.Context, string, []byte) (core.MediaSource, error) {
	return nil, context.Canceled
}
func (stubEngine) Prefetch(context.Context, string, []byte, int64, int64) error { return nil }
