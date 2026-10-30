package arr

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/braedonsaunders/ether/internal/config"
	"github.com/braedonsaunders/ether/internal/core"
)

// DiscoverOptions control Play-time Arr discovery.
type DiscoverOptions struct {
	// EnsureMissing asks Arr to search when the title is known but has no file,
	// then polls briefly for hasFile before giving up.
	EnsureMissing bool
	// EnsureWait is how long to poll after kicking Ensure (default 8s).
	EnsureWait time.Duration
	// EnsurePoll is the poll interval (default 1s).
	EnsurePoll time.Duration
}

// Discover returns a ProtocolLocal release when Arr has a readable file for t.
func Discover(ctx context.Context, t *core.Title, cfg config.Arr, opts DiscoverOptions) []core.Release {
	if !cfg.Enabled || t == nil {
		return nil
	}
	mf, err := Lookup(ctx, t, cfg)
	if err != nil || mf == nil || !mf.HasFile {
		if opts.EnsureMissing {
			if started, _ := Ensure(ctx, t, cfg); started {
				mf = waitForFile(ctx, t, cfg, opts)
			}
		}
		if mf == nil || !mf.HasFile {
			return nil
		}
	}
	local, ok := ResolveLocal(mf.ArrPath, cfg.PathMaps)
	if !ok {
		return nil
	}
	st, err := os.Stat(local)
	if err != nil || !st.Mode().IsRegular() {
		return nil
	}
	name := mf.Title
	if name == "" {
		name = filepath.Base(local)
	}
	return []core.Release{{
		ID:        ReleaseID(mf),
		GUID:      ReleaseID(mf),
		Indexer:   "Arr",
		Name:      name,
		Size:      st.Size(),
		Protocol:  core.ProtocolLocal,
		StreamURL: local, // absolute path; server-only
		Info: core.ReleaseInfo{
			Title:     name,
			Container: filepath.Ext(local),
		},
		Score: 1e9, // prefer over indexer results in the same feed window
	}}
}

func waitForFile(ctx context.Context, t *core.Title, cfg config.Arr, opts DiscoverOptions) *MediaFile {
	wait := opts.EnsureWait
	if wait <= 0 {
		wait = 3 * time.Second
	}
	poll := opts.EnsurePoll
	if poll <= 0 {
		poll = time.Second
	}
	deadline := time.Now().Add(wait)
	for {
		mf, err := Lookup(ctx, t, cfg)
		if err == nil && mf != nil && mf.HasFile {
			return mf
		}
		if time.Now().After(deadline) {
			return nil
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(poll):
		}
	}
}
