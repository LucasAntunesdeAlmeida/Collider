package collider

import (
	"errors"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// fakeWindow stands in for the window's close button: closing reports
// what IsWindowBeingClosed would, handled what SetWindowClosingHandled
// was last told.
type fakeWindow struct{ closing, handled bool }

func useFakeWindow(t *testing.T) *fakeWindow {
	w := &fakeWindow{}
	oldClosed, oldHandled := windowBeingClosed, setWindowClosingHandled
	windowBeingClosed = func() bool { return w.closing }
	setWindowClosingHandled = func(on bool) { w.handled = on }
	t.Cleanup(func() { windowBeingClosed, setWindowClosingHandled = oldClosed, oldHandled })
	return w
}

// closeGame counts frames; its close handler counts calls and saves
// the frame count, the way a game banks a run in progress.
func closeGame() (g *Game, ticks, calls *int) {
	g = New("close-test", 800, 600)
	s := g.Scene("play")
	ticks, calls = new(int), new(int)
	s.OnUpdate(func(float64) { *ticks++ })
	g.OnClose(func() {
		*calls++
		if err := g.Save("progress", *ticks); err != nil {
			panic(err)
		}
	})
	return g, ticks, calls
}

func TestOnCloseRunsOnceThenQuits(t *testing.T) {
	w := useFakeWindow(t)
	g, ticks, calls := closeGame()
	if !w.handled {
		t.Fatal("OnClose should take over the window's close request")
	}
	g.Go("play")
	r := &runner{g}
	for range 3 {
		if err := r.Update(); err != nil {
			t.Fatalf("an open window should run normally, got %v", err)
		}
	}

	w.closing = true // the player clicks X
	if err := r.Update(); !errors.Is(err, ebiten.Termination) {
		t.Fatalf("a close request should end the game, got %v", err)
	}
	if *calls != 1 {
		t.Fatalf("the handler should run once, ran %d times", *calls)
	}
	if *ticks != 3 {
		t.Fatalf("no frame should run once the window is closing, got %d updates", *ticks)
	}
	var saved int
	if !g.Load("progress", &saved) || saved != 3 {
		t.Fatalf("the handler's save should be there: %d", saved)
	}

	for range 3 { // the loop may call Update again before it stops
		if err := r.Update(); !errors.Is(err, ebiten.Termination) {
			t.Fatalf("a closed game should stay ended, got %v", err)
		}
	}
	if *calls != 1 || *ticks != 3 {
		t.Fatalf("nothing should run after the close: calls=%d ticks=%d", *calls, *ticks)
	}
}

func TestOnCloseNotOnQuit(t *testing.T) {
	useFakeWindow(t)
	g, _, calls := closeGame()
	g.Go("play")
	r := &runner{g}
	r.Update()
	g.Quit()
	if err := r.Update(); !errors.Is(err, ebiten.Termination) {
		t.Fatalf("Quit should end the game, got %v", err)
	}
	if *calls != 0 {
		t.Fatal("Quit is the game's own decision: the close handler must not run")
	}
}

func TestOnCloseNil(t *testing.T) {
	w := useFakeWindow(t)
	g, ticks, calls := closeGame()
	g.OnClose(nil)
	if w.handled {
		t.Fatal("OnClose(nil) should give the close button back to the window")
	}
	g.Go("play")
	r := &runner{g}
	w.closing = true
	if err := r.Update(); err != nil {
		t.Fatalf("without a handler the engine should not end the loop itself, got %v", err)
	}
	if *calls != 0 || *ticks != 1 {
		t.Fatalf("calls=%d ticks=%d", *calls, *ticks)
	}
}

func TestOnCloseHeadless(t *testing.T) {
	w := useFakeWindow(t)
	g, _, calls := closeGame()
	g.Headless("play")
	w.closing = true // not a window of this run: never read headless
	for range 5 {
		if obs := g.Step(Action{}); obs.Quit {
			t.Fatal("a headless run has no window to close")
		}
	}
	if *calls != 0 {
		t.Fatal("the close handler never runs headless")
	}
}
