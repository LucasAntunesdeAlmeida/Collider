//go:build !js

package collider

import (
	"os"
	"path/filepath"
)

// saveBaseDir is the folder that holds each game's save folder: the
// user's config directory (%AppData% on Windows, ~/Library/Application
// Support on macOS, ~/.config on Linux). Tests point it elsewhere.
var saveBaseDir = os.UserConfigDir

// savePath is where a key lives: <config dir>/<game title>/<key>.json.
func savePath(title, key string) (string, error) {
	base, err := saveBaseDir()
	if err != nil {
		return "", errNoStorage
	}
	return filepath.Join(base, saveKey(title), key+".json"), nil
}

// writeSave writes atomically: a temp file in the same folder, then a
// rename over the old save, so a crash mid-write never leaves a
// truncated file behind.
func writeSave(title, key string, b []byte) error {
	path, err := savePath(title, key)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, key+".*.tmp")
	if err != nil {
		return err
	}
	_, werr := tmp.Write(b)
	cerr := tmp.Close()
	if werr == nil {
		werr = cerr
	}
	if werr == nil {
		werr = os.Rename(tmp.Name(), path)
	}
	if werr != nil {
		os.Remove(tmp.Name())
	}
	return werr
}

func readSave(title, key string) ([]byte, bool) {
	path, err := savePath(title, key)
	if err != nil {
		return nil, false
	}
	b, err := os.ReadFile(path)
	return b, err == nil
}
