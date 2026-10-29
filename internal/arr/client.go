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
	Title   string `json:"title"`
	ArrPath string `json:"arrPath"`
	HasFile bool   `json:"hasFile"`
	ArrID   int    `json:"arrId,omitempty"`
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
		return &MediaFile{Title: m.Title, ArrPath: m.MovieFile.Path, HasFile: true, ArrID: m.ID}, nil
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
