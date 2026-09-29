package collider

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// quitGame counts frames and calls Quit on frame quitAt, destroying a
// crate in the same frame.
func quitGame(quitAt int) (*Game, *int) {
	g := New("quit-test", 800, 600)
	s := g.Scene("play")
	crate := s.Add(Rect(10, 10, Red).At(50, 50).Tag("crate"))
	ticks := 0
	s.Every(1.0/60, func() {}) // a timer that must stop with the run
	s.OnUpdate(func(float64) {
		ticks++
		if ticks == quitAt {
			crate.Destroy()
			g.Quit()
		}
	})
	return g, &ticks
}

func TestQuitHeadlessEndsTheRun(t *testing.T) {
	g, ticks := quitGame(3)
	g.Headless("play")
	*ticks = 0
	if !g.CanQuit() {
		t.Fatal("headless play can always quit")
	}

	var obs Observation
	for range 3 {
		obs = g.Step(Action{})
	}
	if !obs.Quit {
		t.Fatal("the frame that called Quit should report it")
	}
	if len(obs.Objects) != 0 {
		t.Fatalf("the quitting frame still finishes: its Destroy applies, got %+v", obs.Objects)
	}

	for range 5 {
		obs = g.Step(Action{Keys: []Key{Right}})
	}
	if *ticks != 3 {
		t.Fatalf("Step after Quit should advance nothing, got %d updates", *ticks)
	}
	if !obs.Quit || obs.Scene != "play" {
		t.Fatalf("Step after Quit should return the final observation, got %+v", obs)
	}

	g.Headless("play") // a new run (activating the scene runs a frame)
	if obs = g.Step(Action{}); obs.Quit || *ticks != 5 {
		t.Fatalf("Headless should start a new run: quit=%v, ticks=%d", obs.Quit, *ticks)
	}
}

func TestQuitInObservationJSON(t *testing.T) {
	g, _ := quitGame(2) // Headless runs frame 1 to activate the scene
	g.Headless("play")
	b, _ := json.Marshal(g.Observe())
	if strings.Contains(string(b), "quit") {
		t.Fatalf("a running game should not mention quit: %s", b)
	}
	b, _ = json.Marshal(g.Step(Action{}))
	if !strings.Contains(string(b), `"quit":true`) {
		t.Fatalf("a quit game should say so: %s", b)
	}
}

func TestQuitWindowedEndsAfterTheFrame(t *testing.T) {
	g, ticks := quitGame(2)
	g.Go("play")
	r := &runner{g}
	if !g.CanQuit() {
		t.Fatal("desktop games can quit")
	}

	for frame := 1; frame <= 2; frame++ {
		if err := r.Update(); err != nil {
			t.Fatalf("frame %d should run normally, got %v", frame, err)
		}
	}
	if *ticks != 2 {
		t.Fatalf("the frame that called Quit should finish, got %d updates", *ticks)
	}
	if err := r.Update(); !errors.Is(err, ebiten.Termination) {
		t.Fatalf("the frame after Quit should end the loop, got %v", err)
	}
	if *ticks != 2 {
		t.Fatal("no update should run once the game has quit")
	}
}

func TestQuitDoesNothingOnTheWeb(t *testing.T) {
	old := quitSupported
	quitSupported = false
	t.Cleanup(func() { quitSupported = old })

	g, ticks := quitGame(1)
	g.Go("play")
	r := &runner{g}
	if g.CanQuit() {
		t.Fatal("a browser game cannot quit")
	}
	for range 3 {
		if err := r.Update(); err != nil {
			t.Fatalf("Quit in a browser should be ignored, got %v", err)
		}
	}
	if *ticks != 3 {
		t.Fatalf("the game should keep running, got %d updates", *ticks)
	}

	// Headless runs are simulations: they can quit anywhere.
	h, _ := quitGame(2)
	h.Headless("play")
	if !h.CanQuit() || !h.Step(Action{}).Quit {
		t.Fatal("headless play should quit even where windows cannot")
	}
}
