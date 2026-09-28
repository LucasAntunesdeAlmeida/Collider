//go:build !js

package collider

import (
	"os"
	"path/filepath"
	"testing"
)

type progress struct {
	Best   int      `json:"best"`
	Name   string   `json:"name"`
	Unlock []string `json:"unlock"`
}

// onDisk points saves at a temp folder and turns off the in-memory
// store tests normally get, so the real file backend runs.
func onDisk(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	oldBase, oldTest := saveBaseDir, underTest
	saveBaseDir = func() (string, error) { return dir, nil }
	underTest = func() bool { return false }
	t.Cleanup(func() { saveBaseDir, underTest = oldBase, oldTest })
	return dir
}

func TestSaveLoadRoundtripHeadless(t *testing.T) {
	g := New("save-test", 800, 600)
	g.Scene("play")
	g.Headless("play")

	in := progress{Best: 42, Name: "ada", Unlock: []string{"bow"}}
	if err := g.Save("progress", in); err != nil {
		t.Fatal(err)
	}
	var out progress
	if !g.Load("progress", &out) || out.Best != 42 || out.Name != "ada" || len(out.Unlock) != 1 {
		t.Fatalf("roundtrip failed: %+v", out)
	}

	// A different game keeps its own saves.
	other := New("save-test", 800, 600)
	other.Scene("play")
	other.Headless("play")
	if other.Load("progress", &out) {
		t.Fatal("headless saves are per game, in memory")
	}
}

func TestTestsNeverTouchTheDisk(t *testing.T) {
	dir := t.TempDir()
	old := saveBaseDir
	saveBaseDir = func() (string, error) { return dir, nil }
	t.Cleanup(func() { saveBaseDir = old })

	// Not headless, but under go test: in memory.
	g := New("save-test", 800, 600)
	if err := g.Save("best", 7); err != nil {
		t.Fatal(err)
	}
	var n int
	if !g.Load("best", &n) || n != 7 {
		t.Fatalf("in-memory roundtrip, got %d", n)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Fatalf("a game under test wrote to disk: %v", entries)
	}
}

func TestSaveLoadRoundtripOnDisk(t *testing.T) {
	dir := onDisk(t)
	g := New("Save: Test!", 800, 600)

	if err := g.Save("progress", progress{Best: 9, Unlock: []string{"a", "b"}}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "Save__Test_", "progress.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("save file should be at %s: %v", path, err)
	}

	// Overwrite (atomic rename over the old file), then a fresh game
	// with the same title reads it back: the next run.
	if err := g.Save("progress", map[string]int{"best": 10}); err != nil { // an older save: no name
		t.Fatal(err)
	}
	out := progress{Name: "default"}
	if !New("Save: Test!", 800, 600).Load("progress", &out) {
		t.Fatal("next run should load the save")
	}
	if out.Best != 10 || out.Name != "default" {
		t.Fatalf("loaded %+v; missing fields should keep their defaults", out)
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Fatalf("temp files should not be left behind: %v", entries)
	}
}

func TestLoadMissingAndCorrupt(t *testing.T) {
	dir := onDisk(t)
	g := New("corrupt", 800, 600)

	out := progress{Best: 3}
	if g.Load("nothing", &out) || out.Best != 3 {
		t.Fatal("a missing key should return false and leave v alone")
	}

	os.MkdirAll(filepath.Join(dir, "corrupt"), 0o755)
	os.WriteFile(filepath.Join(dir, "corrupt", "progress.json"), []byte(`{"best": 5, "name": `), 0o644)
	if g.Load("progress", &out) || out.Best != 3 {
		t.Fatalf("corrupt data should return false and leave v alone, got %+v", out)
	}

	// Valid JSON of the wrong shape: a partial decode must not leak in.
	os.WriteFile(filepath.Join(dir, "corrupt", "progress.json"), []byte(`{"best": 5, "name": 12}`), 0o644)
	if g.Load("progress", &out) || out.Best != 3 {
		t.Fatalf("undecodable data should leave v alone, got %+v", out)
	}

	if g.Load("progress", out) {
		t.Fatal("a non-pointer cannot be loaded into")
	}
	if err := g.Save("bad", func() {}); err == nil {
		t.Fatal("an unencodable value should error")
	}
}

func TestSaveKeySanitizing(t *testing.T) {
	for in, want := range map[string]string{
		"best":          "best",
		"slot-1_v2.bak": "slot-1_v2.bak",
		"../../etc":     ".._.._etc",
		`a\b/c:d*e`:     "a_b_c_d_e",
		"":              "_",
		"héllo wörld":   "h_llo_w_rld",
	} {
		if got := saveKey(in); got != want {
			t.Errorf("saveKey(%q) = %q, want %q", in, got, want)
		}
	}

	// A hostile key stays inside the game's folder.
	dir := onDisk(t)
	g := New("keys", 800, 600)
	if err := g.Save("../escape", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "keys", ".._escape.json")); err != nil {
		t.Fatalf("sanitized key should be a file in the game folder: %v", err)
	}
	var n int
	if !g.Load("../escape", &n) || n != 1 {
		t.Fatal("the same key should load back")
	}
}
