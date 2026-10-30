package arr

import (
	"context"

	"github.com/braedonsaunders/ether/internal/core"
	"github.com/braedonsaunders/ether/internal/debrid"
)

// Indexer wraps a debrid indexer so ProtocolLocal releases fetch a local payload.
type Indexer struct {
	*debrid.Indexer
}

// WrapIndexer returns an Indexer that serves Arr local payloads from FetchNZB.
func WrapIndexer(inner *debrid.Indexer) *Indexer {
	return &Indexer{Indexer: inner}
}

// FetchNZB implements core.Indexer.
func (x *Indexer) FetchNZB(ctx context.Context, r core.Release) ([]byte, error) {
	if r.Protocol == core.ProtocolLocal {
		return PayloadFromRelease(r)
	}
	return x.Indexer.FetchNZB(ctx, r)
}

// Ensure Indexer still satisfies core.Indexer.
var _ core.Indexer = (*Indexer)(nil)
