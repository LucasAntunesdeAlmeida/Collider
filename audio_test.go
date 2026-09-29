package collider

import (
	"fmt"
	"strings"
	"testing"
)

// Audio calls are tested without the device: the files do not exist,
// so a call that does any audio work panics loading them, and one that
// stays silent returns quietly.

// audible reports whether fn tried to play audio.
func audible(t *testing.T, fn func()) (played bool) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			if !strings.Contains(fmt.Sprint(r), "cannot open audio file") {
				panic(r)
			}
			played = true
		}
	}()
	fn()
	return false
}

// realRun turns off the go test silence, so a game behaves like one a
// player launched.
func realRun(t *testing.T) {
	t.Helper()
	old := underTest
	underTest = func() bool { return false }
	t.Cleanup(func() { underTest = old })
}

// soundAndMusic plays a sound, then enters a scene with music.
func soundAndMusic(g *Game) func() {
	g.Scene("play").Music("no-such-music.wav")
	return func() {
		g.Sound("no-such-sound.wav")
		g.Go("play")
		g.advance(0)
	}
}

func TestTestsAreSilent(t *testing.T) {
	g := New("audio-test", 800, 600) // windowed, but under go test
	if audible(t, soundAndMusic(g)) || g.musicPlayer != nil {
		t.Fatal("go test must not play audio")
	}
}

func TestAgentRunsAreSilent(t *testing.T) {
	realRun(t)
	cases := map[string]func(t *testing.T, g *Game){
		"headless":  func(_ *testing.T, g *Game) { g.Headless("menu") },
		"autopilot": func(_ *testing.T, g *Game) { g.Autopilot(func(Observation) Action { return Action{} }) },
		"mcp":       func(t *testing.T, _ *Game) { t.Setenv("COLLIDER_AGENT", "mcp") },
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			g := New("audio-test", 800, 600)
			g.Scene("menu")
			play := soundAndMusic(g)
			setup(t, g)
			if audible(t, play) {
				t.Fatal("an agent run must not play audio")
			}
		})
	}
}

func TestPlayerRunsPlayAudio(t *testing.T) {
	realRun(t)
	cases := map[string]func(t *testing.T, g *Game){
		"plain": func(*testing.T, *Game) {},
		// A real window a person watches and plays along with.
		"mcp-window": func(t *testing.T, _ *Game) { t.Setenv("COLLIDER_AGENT", "mcp-window") },
		// COLLIDER_AGENT is ignored by a game that disallows agents.
		"mcp, agents disallowed": func(t *testing.T, g *Game) {
			t.Setenv("COLLIDER_AGENT", "mcp")
			g.DisallowAgents()
		},
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			g := New("audio-test", 800, 600)
			setup(t, g)
			if !audible(t, func() { g.Sound("no-such-sound.wav") }) {
				t.Fatal("Sound should play")
			}
			g.Scene("play").Music("no-such-music.wav")
			if !audible(t, func() { g.Go("play"); g.advance(0) }) {
				t.Fatal("music should play")
			}
		})
	}
}
