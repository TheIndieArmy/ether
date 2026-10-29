package config

import "strings"

// Arr is optional Sonarr/Radarr integration for local-file playback and
// request/search. When Enabled is false, instances are ignored.
type Arr struct {
	Enabled bool `json:"enabled"`
	Radarr  ArrInstance `json:"radarr"`
	Sonarr  ArrInstance `json:"sonarr"`
	// PathMaps rewrite Arr container paths to paths Ether can open.
	PathMaps []ArrPathMap `json:"pathMaps"`
}

// ArrInstance is one Starr app connection.
type ArrInstance struct {
	URL              string `json:"url"`
	APIKey           string `json:"apiKey"`
	QualityProfileID int    `json:"qualityProfileId"`
	RootFolderPath   string `json:"rootFolderPath"`
}

// ArrPathMap maps an Arr-side path prefix to a local prefix Ether can read.
type ArrPathMap struct {
	ArrPrefix   string `json:"arrPrefix"`
	LocalPrefix string `json:"localPrefix"`
}

// Configured reports whether the instance has enough to call the API.
func (a ArrInstance) Configured() bool {
	return strings.TrimSpace(a.URL) != "" && strings.TrimSpace(a.APIKey) != ""
}
