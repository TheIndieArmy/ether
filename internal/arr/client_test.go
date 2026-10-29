package arr

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProbeAndMovieFile(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v3/system/status", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"version": "5.0.0", "appName": "Radarr"})
	})
	mux.HandleFunc("/api/v3/rootfolder", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 1, "path": "/movies"}})
	})
	mux.HandleFunc("/api/v3/qualityprofile", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 2, "name": "HD"}})
	})
	mux.HandleFunc("/api/v3/movie", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("tmdbId") != "550" {
			t.Fatalf("unexpected tmdbId %q", r.URL.Query().Get("tmdbId"))
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"id": 9, "title": "Fight Club", "hasFile": true,
			"movieFile": map[string]any{"path": "/movies/Fight Club.mkv"},
		}})
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	c := New(ts.URL, "key", "radarr")
	st, roots, profiles, err := c.Probe(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if st.Version != "5.0.0" || len(roots) != 1 || roots[0].Path != "/movies" || len(profiles) != 1 {
		t.Fatalf("probe: %+v %+v %+v", st, roots, profiles)
	}
	mf, err := c.MovieFileByTMDB(t.Context(), 550)
	if err != nil || mf == nil || mf.ArrPath == "" || !mf.HasFile {
		t.Fatalf("movie file: %+v %v", mf, err)
	}
}
