package arr

import (
	"context"
	"fmt"

	"github.com/braedonsaunders/ether/internal/config"
	"github.com/braedonsaunders/ether/internal/core"
)

// Lookup finds an Arr media file for t when Arr is enabled and the matching
// instance is configured. Returns nil when there is no hasFile hit.
func Lookup(ctx context.Context, t *core.Title, cfg config.Arr) (*MediaFile, error) {
	if !cfg.Enabled || t == nil {
		return nil, nil
	}
	switch t.Type {
	case core.TypeMovie:
		if !cfg.Radarr.Configured() {
			return nil, nil
		}
		tmdb := int(t.IDs.TMDB)
		if tmdb <= 0 {
			tmdb = int(t.Key.TMDB())
		}
		c := New(cfg.Radarr.URL, cfg.Radarr.APIKey, "radarr")
		return c.MovieFileByTMDB(ctx, tmdb)
	case core.TypeEpisode:
		if !cfg.Sonarr.Configured() {
			return nil, nil
		}
		tvdb, tmdb := t.ShowIDs.TVDB, t.ShowIDs.TMDB
		if tmdb <= 0 {
			tmdb = t.Key.TMDB()
		}
		c := New(cfg.Sonarr.URL, cfg.Sonarr.APIKey, "sonarr")
		return c.EpisodeFileByShow(ctx, tvdb, tmdb, t.Season, t.Episode)
	default:
		return nil, nil
	}
}

// ReleaseID is the stable resolver release id for an Arr media file.
func ReleaseID(mf *MediaFile) string {
	if mf == nil {
		return ""
	}
	if mf.Kind == "sonarr" && mf.EpisodeID > 0 {
		return fmt.Sprintf("arr:sonarr:%d:%d", mf.ArrID, mf.EpisodeID)
	}
	kind := mf.Kind
	if kind == "" {
		kind = "radarr"
	}
	return fmt.Sprintf("arr:%s:%d", kind, mf.ArrID)
}

// IsArrRelease reports whether id was minted by ReleaseID.
func IsArrRelease(id string) bool {
	return len(id) > 4 && id[:4] == "arr:"
}
