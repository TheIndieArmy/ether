package arr

import (
	"context"
	"fmt"
	"os"

	"github.com/braedonsaunders/ether/internal/core"
)

// Engine wraps an inner stream engine and serves ProtocolLocal payloads from disk.
type Engine struct {
	inner  core.Engine
	health core.HealthChecker
}

// WrapEngine returns an Engine that opens local payloads and delegates otherwise.
// inner may also implement core.HealthChecker / ReleaseHealthChecker.
func WrapEngine(inner core.Engine) *Engine {
	e := &Engine{inner: inner}
	if hc, ok := inner.(core.HealthChecker); ok {
		e.health = hc
	}
	return e
}

// Open implements core.Engine.
func (e *Engine) Open(ctx context.Context, releaseID string, nzb []byte) (core.MediaSource, error) {
	if IsLocalPayload(nzb) {
		p, err := DecodeLocalPayload(nzb)
		if err != nil {
			return nil, err
		}
		if p.ReleaseID == "" {
			p.ReleaseID = releaseID
		}
		return openFileSource(p)
	}
	return e.inner.Open(ctx, releaseID, nzb)
}

// Prefetch implements core.Engine. Local files need no prefetch.
func (e *Engine) Prefetch(ctx context.Context, releaseID string, nzb []byte, offset, length int64) error {
	if IsLocalPayload(nzb) {
		return nil
	}
	return e.inner.Prefetch(ctx, releaseID, nzb, offset, length)
}

// DurablePrefetch implements core.DurablePrefetcher when the inner engine does.
func (e *Engine) DurablePrefetch(ctx context.Context, releaseID string, nzb []byte, offset, length int64) error {
	if IsLocalPayload(nzb) {
		return nil
	}
	if d, ok := e.inner.(core.DurablePrefetcher); ok {
		return d.DurablePrefetch(ctx, releaseID, nzb, offset, length)
	}
	return e.inner.Prefetch(ctx, releaseID, nzb, offset, length)
}

// Check implements core.HealthChecker.
func (e *Engine) Check(ctx context.Context, nzb []byte) (bool, error) {
	return e.CheckRelease(ctx, "", nzb)
}

// CheckRelease implements core.ReleaseHealthChecker.
func (e *Engine) CheckRelease(ctx context.Context, releaseID string, nzb []byte) (bool, error) {
	if IsLocalPayload(nzb) {
		p, err := DecodeLocalPayload(nzb)
		if err != nil {
			return false, err
		}
		st, err := os.Stat(p.Path)
		if err != nil {
			return false, err
		}
		if !st.Mode().IsRegular() {
			return false, fmt.Errorf("arr: not a regular file")
		}
		return true, nil
	}
	if e.health == nil {
		return true, nil
	}
	return core.CheckReleaseHealth(ctx, e.health, releaseID, nzb)
}
