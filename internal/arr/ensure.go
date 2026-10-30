package arr

import (
	"context"
	"strconv"

	"github.com/braedonsaunders/ether/internal/config"
	"github.com/braedonsaunders/ether/internal/core"
)

// Ensure asks Arr to search for t when it is already in the library but
// missing a file. It does not add new titles. started is true when a search
// command was sent so Discover can poll briefly for hasFile.
func Ensure(ctx context.Context, t *core.Title, cfg config.Arr) (started bool, err error) {
	if !cfg.Enabled || t == nil {
		return false, nil
	}
	switch t.Type {
	case core.TypeMovie:
		if !cfg.Radarr.Configured() {
			return false, nil
		}
		return ensureMovie(ctx, New(cfg.Radarr.URL, cfg.Radarr.APIKey, "radarr"), t)
	case core.TypeEpisode:
		if !cfg.Sonarr.Configured() {
			return false, nil
		}
		return ensureEpisode(ctx, New(cfg.Sonarr.URL, cfg.Sonarr.APIKey, "sonarr"), t)
	default:
		return false, nil
	}
}

func ensureMovie(ctx context.Context, c *Client, t *core.Title) (bool, error) {
	tmdb := int(t.IDs.TMDB)
	if tmdb <= 0 {
		tmdb = int(t.Key.TMDB())
	}
	if tmdb <= 0 {
		return false, nil
	}
	mf, err := c.MovieFileByTMDB(ctx, tmdb)
	if err != nil {
		return false, err
	}
	if mf != nil && mf.HasFile {
		return false, nil
	}
	movieID, err := c.findMovieID(ctx, tmdb)
	if err != nil || movieID == 0 {
		return false, err
	}
	if err := c.postJSON(ctx, "/api/v3/command", map[string]any{
		"name":     "MoviesSearch",
		"movieIds": []int{movieID},
	}, nil); err != nil {
		return false, err
	}
	return true, nil
}

func ensureEpisode(ctx context.Context, c *Client, t *core.Title) (bool, error) {
	tvdb, tmdb := t.ShowIDs.TVDB, t.ShowIDs.TMDB
	if tmdb <= 0 {
		tmdb = t.Key.TMDB()
	}
	seriesID, _, err := c.findSeriesID(ctx, tvdb, tmdb)
	if err != nil || seriesID == 0 {
		return false, err
	}
	mf, err := c.EpisodeFileByShow(ctx, tvdb, tmdb, t.Season, t.Episode)
	if err != nil {
		return false, err
	}
	if mf != nil && mf.HasFile {
		return false, nil
	}
	epID := 0
	if mf != nil {
		epID = mf.EpisodeID
	}
	if epID == 0 {
		epID, err = c.findEpisodeID(ctx, seriesID, t.Season, t.Episode)
		if err != nil || epID == 0 {
			return false, err
		}
	}
	if err := c.postJSON(ctx, "/api/v3/command", map[string]any{
		"name":       "EpisodeSearch",
		"episodeIds": []int{epID},
	}, nil); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Client) postJSON(ctx context.Context, path string, body, dest any) error {
	return c.doJSON(ctx, "POST", path, body, dest)
}

func (c *Client) findMovieID(ctx context.Context, tmdbID int) (int, error) {
	var movies []struct {
		ID int `json:"id"`
	}
	if err := c.getJSON(ctx, "/api/v3/movie?tmdbId="+strconv.Itoa(tmdbID), &movies); err != nil {
		return 0, err
	}
	if len(movies) == 0 {
		return 0, nil
	}
	return movies[0].ID, nil
}

func (c *Client) findEpisodeID(ctx context.Context, seriesID, season, episode int) (int, error) {
	var eps []struct {
		ID            int `json:"id"`
		SeasonNumber  int `json:"seasonNumber"`
		EpisodeNumber int `json:"episodeNumber"`
	}
	q := "/api/v3/episode?seriesId=" + strconv.Itoa(seriesID) +
		"&seasonNumber=" + strconv.Itoa(season)
	if err := c.getJSON(ctx, q, &eps); err != nil {
		return 0, err
	}
	for _, ep := range eps {
		if ep.EpisodeNumber == episode {
			return ep.ID, nil
		}
	}
	return 0, nil
}
