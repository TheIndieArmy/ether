package resolver

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/braedonsaunders/ether/internal/core"
	"github.com/braedonsaunders/ether/internal/releaseparse"
)

func TestFeedValidationNoSearchAndNoTitleNegative(t *testing.T) {
	idx := &fakeIndexer{}
	count := &atomic.Int64{}
	r := rel("good", "The.Signal.1999.1080p.BluRay.x264-GRP", 8000)
	r.Info = releaseparse.Parse(r.Name)
	s := New(idx, fakeHealth{checks: count}, nil, testCfg(), Options{})
	defer s.Close()
	if e := s.VerifyFeedCandidate(t.Context(), matrixTitle(), r); e != nil {
		t.Fatal(e)
	}
	if idx.searches.Load() != 0 || count.Load() != 1 {
		t.Fatal("feed validation searched or skipped health")
	}
	if e := s.VerifyFeedCandidate(t.Context(), matrixTitle(), r); e != nil {
		t.Fatal(e)
	}
	if count.Load() != 2 {
		t.Fatal("refresh used old health cache")
	}
	s2 := New(idx, fakeHealth{unhealthy: map[string]bool{"good": true}}, nil, testCfg(), Options{})
	defer s2.Close()
	if e := s2.VerifyFeedCandidate(t.Context(), matrixTitle(), r); !errors.Is(e, core.ErrSourceUnavailable) {
		t.Fatal(e)
	}
	if known, _ := s2.Available(t.Context(), matrixTitle().Key); known {
		t.Fatal("one failed source made title unavailable")
	}
}
func TestDiscoveredDescriptorAvoidsSearchAndCachedHealthDoesNotRenewObservation(t *testing.T) {
	idx := &fakeIndexer{}
	var observed atomic.Int32
	r := rel("known", "The.Signal.1999.1080p.BluRay.x264-GRP", 8000)
	r.Info = releaseparse.Parse(r.Name)
	s := New(idx, fakeHealth{}, nil, testCfg(), Options{Discovered: func(_ context.Context, _ *core.Title) []core.Release { return []core.Release{r} }, Observed: func(_ context.Context, _ *core.Title, _ core.Release) { observed.Add(1) }})
	defer s.Close()
	for i := 0; i < 2; i++ {
		res, e := s.Resolve(t.Context(), core.ResolveRequest{Title: matrixTitle(), Client: webCaps()})
		if e != nil || res == nil {
			t.Fatal(res, e)
		}
	}
	if idx.searches.Load() != 0 || observed.Load() != 1 {
		t.Fatalf("searches %d observations %d", idx.searches.Load(), observed.Load())
	}
	for _, trace := range s.Traces() {
		if len(trace.Discovery) != 1 || trace.Discovery[0].Path != "feed" || trace.Discovery[0].Count != 1 {
			t.Fatalf("feed trace: %+v", trace.Discovery)
		}
	}
}

// A sticky Prefer (e.g. last Usenet pick) must not hide Arr hasFile discovery.
func TestDiscoveredLocalStillWinsWhenPreferSet(t *testing.T) {
	idx := &fakeIndexer{}
	local := core.Release{
		ID: "arr:radarr:1", Name: "Local File", Protocol: core.ProtocolLocal,
		StreamURL: "/tmp/unused", Score: 1e9, Size: 8 << 30,
		Info: releaseparse.Parse("Movie.2020.2160p.BluRay.x265"),
	}
	s := New(idx, fakeHealth{}, nil, testCfg(), Options{
		Discovered: func(context.Context, *core.Title) []core.Release { return []core.Release{local} },
	})
	defer s.Close()
	res, err := s.Resolve(t.Context(), core.ResolveRequest{
		Title: matrixTitle(), Client: webCaps(), Prefer: "some-usenet-id",
	})
	if err != nil || res == nil {
		t.Fatal(res, err)
	}
	if res.Release.ID != local.ID || res.Release.Protocol != core.ProtocolLocal {
		t.Fatalf("prefer hid local discovery: %+v", res.Release)
	}
	if idx.searches.Load() != 0 {
		t.Fatalf("searched despite local discovery: %d", idx.searches.Load())
	}
}
