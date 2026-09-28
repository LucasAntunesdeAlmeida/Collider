package collider

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
)

// Save stores v as JSON under key, for the next run of the game: best
// scores, unlocks, settings. On desktop each key is a file in the
// user's config directory (under a folder named after the game title);
// in the browser it is an entry in localStorage. Games run by agents
// or tests (Headless, Autopilot, COLLIDER_AGENT=mcp, go test) keep
// saves in memory instead, so bots and CI never touch the player's
// data.
//
// Keys are simple names like "best" or "settings"; characters other
// than letters, digits, '-', '_' and '.' are replaced.
func (g *Game) Save(key string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if g.saveInMemory() {
		if g.memSaves == nil {
			g.memSaves = map[string][]byte{}
		}
		g.memSaves[saveKey(key)] = b
		return nil
	}
	return writeSave(g.title, saveKey(key), b)
}

// Load reads the value saved under key into v, which must be a
// pointer, and reports whether it did. A missing key (the first run)
// or data that does not decode into v returns false and leaves v
// untouched, so defaults set before Load survive. Fields missing from
// an older save keep the values v already had.
func (g *Game) Load(key string, v any) bool {
	var b []byte
	var ok bool
	if g.saveInMemory() {
		b, ok = g.memSaves[saveKey(key)]
	} else {
		b, ok = readSave(g.title, saveKey(key))
	}
	if !ok {
		return false
	}
	dst := reflect.ValueOf(v)
	if dst.Kind() != reflect.Pointer || dst.IsNil() {
		return false
	}
	// Decode into a copy so a half-decoded value never reaches v.
	tmp := reflect.New(dst.Elem().Type())
	tmp.Elem().Set(dst.Elem())
	if err := json.Unmarshal(b, tmp.Interface()); err != nil {
		return false
	}
	dst.Elem().Set(tmp.Elem())
	return true
}

// saveInMemory reports whether saves stay in memory: agent and test
// runs, including a game whose setup calls Load before Run turns it
// into an MCP server.
func (g *Game) saveInMemory() bool {
	return g.headless || g.pilot != nil || underTest() ||
		(os.Getenv("COLLIDER_AGENT") == "mcp" && !g.agentsOff)
}

// saveKey makes a key safe as a file name and a storage key.
func saveKey(key string) string {
	if key == "" {
		return "_"
	}
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9',
			r == '-', r == '_', r == '.':
			return r
		}
		return '_'
	}, key)
}

// underTest is testing.Testing, swappable so the engine's own tests can
// exercise the real storage backend.
var underTest = testing.Testing

var errNoStorage = errors.New("collider: save storage is unavailable")
