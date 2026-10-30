package arr

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/braedonsaunders/ether/internal/config"
)

func TestRewriteLongestPrefix(t *testing.T) {
	maps := []config.ArrPathMap{
		{ArrPrefix: "/data", LocalPrefix: "/mnt/data"},
		{ArrPrefix: "/data/movies", LocalPrefix: "/mnt/movies"},
	}
	got, ok := Rewrite("/data/movies/Foo/foo.mkv", maps)
	if !ok || got != filepath.Join("/mnt/movies", "Foo", "foo.mkv") {
		t.Fatalf("got %q ok=%v", got, ok)
	}
}

func TestRewriteNoMapUsesExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.mkv")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok := Rewrite(path, nil)
	if !ok || got != path {
		t.Fatalf("got %q ok=%v", got, ok)
	}
	if _, ok := Rewrite(filepath.Join(dir, "missing.mkv"), nil); ok {
		t.Fatal("expected miss")
	}
}

func TestResolveLocal(t *testing.T) {
	dir := t.TempDir()
	arrRoot := filepath.Join(dir, "arr")
	localRoot := filepath.Join(dir, "local")
	if err := os.MkdirAll(filepath.Join(localRoot, "Show"), 0o755); err != nil {
		t.Fatal(err)
	}
	localFile := filepath.Join(localRoot, "Show", "ep.mkv")
	if err := os.WriteFile(localFile, []byte("vid"), 0o644); err != nil {
		t.Fatal(err)
	}
	maps := []config.ArrPathMap{{ArrPrefix: arrRoot, LocalPrefix: localRoot}}
	got, ok := ResolveLocal(filepath.Join(arrRoot, "Show", "ep.mkv"), maps)
	if !ok || got != localFile {
		t.Fatalf("got %q ok=%v want %q", got, ok, localFile)
	}
}
