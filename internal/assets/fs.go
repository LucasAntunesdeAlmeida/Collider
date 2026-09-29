package assets

import (
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

// The asset filesystem is process-wide immutable data, like fonts and
// the audio context: set once at startup (UseAssets), read forever.
var (
	fsMu sync.Mutex
	fsys fs.FS // nil = read from disk
)

// SetFS routes all asset loading (sprites, sounds, music, fonts)
// through an embedded or custom filesystem instead of the disk.
func SetFS(f fs.FS) {
	fsMu.Lock()
	defer fsMu.Unlock()
	fsys = f
}

// Stat reports whether an asset exists (nil) without reading it.
func Stat(path string) error {
	fsMu.Lock()
	f := fsys
	fsMu.Unlock()
	if f != nil {
		_, err := fs.Stat(f, filepath.ToSlash(path))
		return err
	}
	_, err := os.Stat(path)
	return err
}

// ReadFile reads an asset from the configured filesystem, or from disk
// when none is set. Paths always use forward slashes in fs.FS.
func ReadFile(path string) ([]byte, error) {
	fsMu.Lock()
	f := fsys
	fsMu.Unlock()
	if f != nil {
		return fs.ReadFile(f, filepath.ToSlash(path))
	}
	return os.ReadFile(path)
}
