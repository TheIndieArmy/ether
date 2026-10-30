package arr

import (
	"fmt"
	"os"
)

// fileSource is a core.MediaSource over a local file.
type fileSource struct {
	f    *os.File
	size int64
	name string
	id   string
}

func (s *fileSource) ReadAt(p []byte, off int64) (int, error) { return s.f.ReadAt(p, off) }
func (s *fileSource) Size() int64                             { return s.size }
func (s *fileSource) Name() string                            { return s.name }
func (s *fileSource) ReleaseID() string                       { return s.id }
func (s *fileSource) Close() error                            { return s.f.Close() }
func (s *fileSource) Origin() string                          { return "Arr" }

func openFileSource(p LocalPayload) (*fileSource, error) {
	f, err := os.Open(p.Path)
	if err != nil {
		return nil, fmt.Errorf("arr: open %s: %w", p.Path, err)
	}
	st, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	size := p.Size
	if size <= 0 {
		size = st.Size()
	}
	name := p.Name
	if name == "" {
		name = st.Name()
	}
	return &fileSource{f: f, size: size, name: name, id: p.ReleaseID}, nil
}
