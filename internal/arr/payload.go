package arr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/braedonsaunders/ether/internal/core"
)

const payloadMagic = "ETHER-LOCAL/1\n"

// LocalPayload identifies a readable file on this host.
type LocalPayload struct {
	ReleaseID string `json:"releaseId"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	Name      string `json:"name,omitempty"`
}

// IsLocalPayload reports whether b is an encoded LocalPayload.
func IsLocalPayload(b []byte) bool {
	return bytes.HasPrefix(b, []byte(payloadMagic))
}

// Encode serializes p with the local payload prefix.
func (p LocalPayload) Encode() []byte {
	body, _ := json.Marshal(p)
	return append([]byte(payloadMagic), body...)
}

// DecodeLocalPayload parses an encoded LocalPayload.
func DecodeLocalPayload(b []byte) (LocalPayload, error) {
	if !IsLocalPayload(b) {
		return LocalPayload{}, errors.New("arr: not a local payload")
	}
	var p LocalPayload
	if err := json.Unmarshal(b[len(payloadMagic):], &p); err != nil {
		return LocalPayload{}, fmt.Errorf("arr: corrupt local payload: %w", err)
	}
	if p.Path == "" {
		return LocalPayload{}, errors.New("arr: local payload missing path")
	}
	return p, nil
}

// PayloadFromPath builds a LocalPayload for an absolute file path.
func PayloadFromPath(releaseID, path string) (LocalPayload, error) {
	st, err := os.Stat(path)
	if err != nil {
		return LocalPayload{}, err
	}
	if !st.Mode().IsRegular() {
		return LocalPayload{}, fmt.Errorf("arr: not a regular file: %s", path)
	}
	return LocalPayload{
		ReleaseID: releaseID,
		Path:      path,
		Size:      st.Size(),
		Name:      filepath.Base(path),
	}, nil
}

// PayloadFromRelease builds a local payload from a ProtocolLocal release.
// The absolute path is carried in StreamURL (server-only).
func PayloadFromRelease(r core.Release) ([]byte, error) {
	if r.Protocol != core.ProtocolLocal {
		return nil, fmt.Errorf("arr: release %q is not local", r.ID)
	}
	path := r.StreamURL
	if path == "" {
		return nil, fmt.Errorf("arr: release %q has no local path", r.ID)
	}
	p, err := PayloadFromPath(r.ID, path)
	if err != nil {
		return nil, err
	}
	if r.Name != "" {
		p.Name = filepath.Base(path)
	}
	return p.Encode(), nil
}
