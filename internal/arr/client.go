package arr

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Client talks to a Radarr or Sonarr instance.
type Client struct {
	BaseURL string
	APIKey  string
	Kind    string // "radarr" or "sonarr"
	HTTP    *http.Client
}

// RootFolder is an Arr root folder.
type RootFolder struct {
	ID   int    `json:"id"`
	Path string `json:"path"`
}

// QualityProfile is an Arr quality profile.
type QualityProfile struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Status is the system status response summary.
type Status struct {
	Version string `json:"version"`
	AppName string `json:"appName"`
}

// MediaFile is a resolvable Arr media file path.
type MediaFile struct {
	Title     string `json:"title"`
	ArrPath   string `json:"arrPath"`
	HasFile   bool   `json:"hasFile"`
	ArrID     int    `json:"arrId,omitempty"`
	Kind      string `json:"kind,omitempty"` // "radarr" or "sonarr"
	EpisodeID int    `json:"episodeId,omitempty"`
}

// New builds a client. kind is "radarr" or "sonarr".
func New(baseURL, apiKey, kind string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		APIKey:  strings.TrimSpace(apiKey),
		Kind:    strings.ToLower(strings.TrimSpace(kind)),
		HTTP:    &http.Client{Timeout: 15 * time.Second},
	}
}

// TestConnection hits /api/v3/system/status.
func (c *Client) TestConnection(ctx context.Context) (Status, error) {
	var st Status
	if err := c.getJSON(ctx, "/api/v3/system/status", &st); err != nil {
		return st, err
	}
	if st.AppName == "" {
		st.AppName = c.Kind
	}
	return st, nil
}

// ListRootFolders returns configured root folders.
func (c *Client) ListRootFolders(ctx context.Context) ([]RootFolder, error) {
	var raw []struct {
		ID   int    `json:"id"`
		Path string `json:"path"`
	}
	if err := c.getJSON(ctx, "/api/v3/rootfolder", &raw); err != nil {
		return nil, err
	}
	out := make([]RootFolder, 0, len(raw))
	for _, r := range raw {
		out = append(out, RootFolder{ID: r.ID, Path: r.Path})
	}
	return out, nil
}

// ListQualityProfiles returns quality profiles.
func (c *Client) ListQualityProfiles(ctx context.Context) ([]QualityProfile, error) {
	var raw []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if err := c.getJSON(ctx, "/api/v3/qualityprofile", &raw); err != nil {
		return nil, err
	}
	out := make([]QualityProfile, 0, len(raw))
	for _, r := range raw {
		out = append(out, QualityProfile{ID: r.ID, Name: r.Name})
	}
	return out, nil
}

// MovieFileByTMDB returns the Radarr movie file for a TMDB id when hasFile.
func (c *Client) MovieFileByTMDB(ctx context.Context, tmdbID int) (*MediaFile, error) {
	if c.Kind != "radarr" {
		return nil, fmt.Errorf("MovieFileByTMDB requires radarr")
	}
	if tmdbID <= 0 {
		return nil, nil
	}
	var movies []struct {
		ID        int    `json:"id"`
		Title     string `json:"title"`
		HasFile   bool   `json:"hasFile"`
		MovieFile *struct {
			Path string `json:"path"`
		} `json:"movieFile"`
	}
	if err := c.getJSON(ctx, "/api/v3/movie?tmdbId="+strconv.Itoa(tmdbID), &movies); err != nil {
		return nil, err
	}
	for _, m := range movies {
		if !m.HasFile || m.MovieFile == nil || m.MovieFile.Path == "" {
			continue
		}
		return &MediaFile{Title: m.Title, ArrPath: m.MovieFile.Path, HasFile: true, ArrID: m.ID, Kind: "radarr"}, nil
	}
	return nil, nil
}

// findSeriesID looks up a Sonarr series by TVDB and/or TMDB id.
// okTitle is unused by callers today; kept for future matching hints.
func (c *Client) findSeriesID(ctx context.Context, tvdb, tmdb int64) (seriesID int, title string, err error) {
	if c.Kind != "sonarr" {
		return 0, "", fmt.Errorf("findSeriesID requires sonarr")
	}
	try := func(path string) (int, string, error) {
		var series []struct {
			ID    int    `json:"id"`
			Title string `json:"title"`
		}
		if err := c.getJSON(ctx, path, &series); err != nil {
			return 0, "", err
		}
		if len(series) == 0 {
			return 0, "", nil
		}
		return series[0].ID, series[0].Title, nil
	}
	if tvdb > 0 {
		if id, title, err := try("/api/v3/series?tvdbId=" + strconv.FormatInt(tvdb, 10)); err != nil || id > 0 {
			return id, title, err
		}
	}
	if tmdb > 0 {
		return try("/api/v3/series?tmdbId=" + strconv.FormatInt(tmdb, 10))
	}
	return 0, "", nil
}

// EpisodeFileByShow returns the Sonarr episode file for a show episode when hasFile.
func (c *Client) EpisodeFileByShow(ctx context.Context, tvdb, tmdb int64, season, episode int) (*MediaFile, error) {
	if c.Kind != "sonarr" {
		return nil, fmt.Errorf("EpisodeFileByShow requires sonarr")
	}
	if season <= 0 || episode <= 0 {
		return nil, nil
	}
	seriesID, seriesTitle, err := c.findSeriesID(ctx, tvdb, tmdb)
	if err != nil || seriesID == 0 {
		return nil, err
	}
	var eps []struct {
		ID            int    `json:"id"`
		Title         string `json:"title"`
		SeasonNumber  int    `json:"seasonNumber"`
		EpisodeNumber int    `json:"episodeNumber"`
		HasFile       bool   `json:"hasFile"`
		EpisodeFile   *struct {
			Path string `json:"path"`
		} `json:"episodeFile"`
	}
	q := "/api/v3/episode?seriesId=" + strconv.Itoa(seriesID) +
		"&seasonNumber=" + strconv.Itoa(season)
	if err := c.getJSON(ctx, q, &eps); err != nil {
		return nil, err
	}
	for _, ep := range eps {
		if ep.EpisodeNumber != episode || !ep.HasFile || ep.EpisodeFile == nil || ep.EpisodeFile.Path == "" {
			continue
		}
		title := ep.Title
		if title == "" {
			title = seriesTitle
		}
		return &MediaFile{
			Title: title, ArrPath: ep.EpisodeFile.Path, HasFile: true,
			ArrID: seriesID, Kind: "sonarr", EpisodeID: ep.ID,
		}, nil
	}
	return nil, nil
}

// Probe gathers status, root folders and quality profiles after a successful connect.
func (c *Client) Probe(ctx context.Context) (Status, []RootFolder, []QualityProfile, error) {
	st, err := c.TestConnection(ctx)
	if err != nil {
		return st, nil, nil, err
	}
	roots, err := c.ListRootFolders(ctx)
	if err != nil {
		return st, nil, nil, err
	}
	profiles, err := c.ListQualityProfiles(ctx)
	if err != nil {
		return st, roots, nil, err
	}
	return st, roots, profiles, nil
}
