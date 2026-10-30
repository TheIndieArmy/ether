package arr

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/braedonsaunders/ether/internal/config"
)

// Rewrite maps an Arr-side path to a path Ether can open using the longest
// matching ArrPrefix. When no map matches, ok is true only if arrPath already
// exists on this host.
func Rewrite(arrPath string, maps []config.ArrPathMap) (local string, ok bool) {
	arrPath = filepath.Clean(strings.TrimSpace(arrPath))
	if arrPath == "" || arrPath == "." {
		return "", false
	}
	bestLen := -1
	bestLocal := ""
	for _, m := range maps {
		prefix := filepath.Clean(strings.TrimSpace(m.ArrPrefix))
		repl := strings.TrimSpace(m.LocalPrefix)
		if prefix == "" || prefix == "." || repl == "" {
			continue
		}
		if arrPath == prefix || strings.HasPrefix(arrPath, prefix+string(filepath.Separator)) ||
			// Also accept slash-normalized prefixes so "/data/movies" matches
			// Arr paths written with either separator style.
			hasPathPrefix(arrPath, prefix) {
			if len(prefix) < bestLen {
				continue
			}
			bestLen = len(prefix)
			suffix := strings.TrimPrefix(normalizeSlashes(arrPath), normalizeSlashes(prefix))
			suffix = strings.TrimPrefix(suffix, "/")
			bestLocal = filepath.Join(filepath.Clean(repl), filepath.FromSlash(suffix))
		}
	}
	if bestLen >= 0 {
		return bestLocal, true
	}
	if fileExists(arrPath) {
		return arrPath, true
	}
	return "", false
}

func normalizeSlashes(p string) string {
	return strings.TrimRight(strings.ReplaceAll(p, "\\", "/"), "/")
}

func hasPathPrefix(path, prefix string) bool {
	p := normalizeSlashes(path)
	pre := normalizeSlashes(prefix)
	return p == pre || strings.HasPrefix(p, pre+"/")
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.Mode().IsRegular()
}

// ResolveLocal rewrites arrPath and confirms the result is a readable file.
func ResolveLocal(arrPath string, maps []config.ArrPathMap) (local string, ok bool) {
	local, ok = Rewrite(arrPath, maps)
	if !ok {
		return "", false
	}
	if !fileExists(local) {
		return "", false
	}
	return local, true
}
