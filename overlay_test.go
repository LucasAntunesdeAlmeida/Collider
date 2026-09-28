package collider

import (
	"encoding/json"
	"strings"
	"testing"
)

// overlayGame is a play scene with a ticking timer, a moving object and
// a click target, plus a "pause" scene with its own timer, update and
// click target at the same spot.
type overlayGame struct {
	g                        *Game
	mover                    *Object
	playTicks, pauseTicks    int
	playUpdates, pauseFrames int
	playClicks, pauseClicks  int
	playSceneClicks          int
	pauseHits                int
}

func newOverlayGame() *overlayGame {
	og := &overlayGame{}
	g := New("overlay-test", 800, 600)
	og.g = g

	play := g.Scene("play")
	og.mover = play.Add(Rect(10, 10, Red).At(100, 100).Tag("player"))
	og.mover.Vx = 60
	play.Every(0.1, func() { og.playTicks++ })
	play.OnUpdate(func(float64) { og.playUpdates++ })
	play.OnClick(func(x, y float64) { og.playSceneClicks++ })
	play.Add(Rect(100, 100, Blue).At(400, 300)).OnClick(func() { og.playClicks++ })

	pause := g.Scene("pause")
	pause.Every(0.1, func() { og.pauseTicks++ })
	pause.OnUpdate(func(float64) { og.pauseFrames++ })
	pause.Add(Rect(100, 100, Green).At(400, 300).Tag("resume")).OnClick(func() { og.pauseClicks++ })
	// Two overlapping objects: the collision fires once the overlay runs.
	pause.Add(Rect(10, 10, Red).At(700, 500).Tag("a"))
	pause.Add(Rect(10, 10, Red).At(705, 500).Tag("b"))
	pause.OnCollision("a", "b", func(a, b *Object) { og.pauseHits++ })

	g.Scene("over").Add(Text("over").At(400, 300))
	return og
}

func TestOverlayFreezesTheSceneUnder(t *testing.T) {
	og := newOverlayGame()
	g := og.g
	g.Headless("play")
	for range 30 {
		g.Step(Action{})
	}
	ticks, updates, x := og.playTicks, og.playUpdates, og.mover.X
	if ticks == 0 || x <= 100 {
		t.Fatalf("play should run before the overlay: ticks %d x %v", ticks, x)
	}

	g.Overlay("pause")
	for range 60 {
		g.Step(Action{})
	}
	if og.playTicks != ticks || og.playUpdates != updates || og.mover.X != x {
		t.Fatalf("scene under the overlay should be frozen: ticks %d->%d updates %d->%d x %v->%v",
			ticks, og.playTicks, updates, og.playUpdates, x, og.mover.X)
	}
	if og.pauseTicks < 9 || og.pauseFrames != 60 {
		t.Fatalf("overlay timers and updates should run: ticks %d frames %d", og.pauseTicks, og.pauseFrames)
	}
	if og.pauseHits != 1 {
		t.Fatalf("overlay collisions should fire once, got %d", og.pauseHits)
	}

	// Closing resumes exactly where it froze: the timer did not advance.
	g.CloseOverlay()
	g.Step(Action{})
	if og.playUpdates != updates+1 || og.mover.X <= x {
		t.Fatal("play should resume after CloseOverlay")
	}
	if og.pauseFrames != 60 {
		t.Fatal("a closed overlay should not update")
	}
}

func TestOverlayIsDeferredToNextFrame(t *testing.T) {
	og := newOverlayGame()
	g := og.g
	g.Headless("play")

	// Opened from inside the play scene's update: that frame belongs to
	// play; the overlay first runs on the next one.
	opened := false
	g.Scene("play").OnUpdate(func(float64) {
		if !opened {
			opened = true
			g.Overlay("pause")
		}
	})
	obs := g.Step(Action{})
	if obs.Overlay != "" || og.pauseFrames != 0 {
		t.Fatalf("overlay should not run on the frame it was opened: %q, %d frames", obs.Overlay, og.pauseFrames)
	}
	updates := og.playUpdates
	obs = g.Step(Action{})
	if obs.Overlay != "pause" || og.pauseFrames != 1 || og.playUpdates != updates {
		t.Fatalf("overlay should take over next frame: %q, %d frames, play %d->%d",
			obs.Overlay, og.pauseFrames, updates, og.playUpdates)
	}

	// Closed from inside the overlay: the rest of that frame is still
	// the overlay's, play resumes on the next.
	g.Scene("pause").OnUpdate(func(float64) { g.CloseOverlay() })
	obs = g.Step(Action{})
	if obs.Overlay != "pause" || og.playUpdates != updates {
		t.Fatal("closing should take effect next frame, not this one")
	}
	obs = g.Step(Action{})
	if obs.Overlay != "" || og.playUpdates != updates+1 {
		t.Fatalf("play should resume the frame after closing: %q, updates %d", obs.Overlay, og.playUpdates)
	}
}

func TestOverlayTakesTheClicks(t *testing.T) {
	og := newOverlayGame()
	g := og.g
	g.Headless("play")

	g.Step(Action{MouseX: 400, MouseY: 300, Click: true})
	if og.playClicks != 1 || og.playSceneClicks != 1 {
		t.Fatal("play should take clicks with no overlay")
	}

	g.Overlay("pause")
	g.Step(Action{})
	g.Step(Action{MouseX: 400, MouseY: 300, Click: true})
	if og.pauseClicks != 1 {
		t.Fatalf("the overlay's button should take the click, got %d", og.pauseClicks)
	}
	if og.playClicks != 1 || og.playSceneClicks != 1 {
		t.Fatal("clicks must not reach the frozen scene under the overlay")
	}
}

func TestGoAndRestartCloseTheOverlay(t *testing.T) {
	og := newOverlayGame()
	g := og.g
	g.Headless("play")
	g.Overlay("pause")
	g.Step(Action{})

	g.Go("over")
	obs := g.Step(Action{})
	if obs.Scene != "over" || obs.Overlay != "" {
		t.Fatalf("Go should switch scenes and close the overlay: %q %q", obs.Scene, obs.Overlay)
	}

	g.Go("play")
	g.Step(Action{})
	g.Overlay("pause")
	g.Step(Action{})
	g.Restart("play")
	obs = g.Step(Action{})
	if obs.Scene != "play" || obs.Overlay != "" {
		t.Fatalf("Restart should close the overlay: %q %q", obs.Scene, obs.Overlay)
	}
}

func TestOverlayReplacesAndKeepsState(t *testing.T) {
	og := newOverlayGame()
	g := og.g
	other := 0
	g.Scene("shop").OnUpdate(func(float64) { other++ })
	g.Headless("play")

	g.Overlay("pause")
	g.Step(Action{})
	g.Step(Action{})
	g.Overlay("shop")
	obs := g.Step(Action{})
	if obs.Overlay != "shop" || other != 1 || og.pauseFrames != 2 {
		t.Fatalf("opening an overlay should replace the open one: %q shop %d pause %d",
			obs.Overlay, other, og.pauseFrames)
	}

	// Reopened, the pause overlay continues where it was: no implicit
	// restart, its timers keep their progress.
	g.Overlay("pause")
	g.Step(Action{})
	if og.pauseFrames != 3 {
		t.Fatalf("a reopened overlay should keep its state, frames %d", og.pauseFrames)
	}

	// A scene cannot overlay itself.
	g.Overlay("play")
	if obs := g.Step(Action{}); obs.Overlay != "" {
		t.Fatalf("the current scene as its own overlay should be ignored, got %q", obs.Overlay)
	}
	g.CloseOverlay() // with nothing open: harmless
	g.Step(Action{})
}

func TestOverlayObservation(t *testing.T) {
	og := newOverlayGame()
	g := og.g
	g.Headless("play")

	before := g.Observe()
	b, _ := json.Marshal(before)
	if strings.Contains(string(b), `"overlay"`) {
		t.Fatalf("no overlay should leave the field out: %s", b)
	}

	g.Overlay("pause")
	obs := g.Step(Action{})
	if obs.Scene != "play" || obs.Overlay != "pause" {
		t.Fatalf("observation should name both: %q %q", obs.Scene, obs.Overlay)
	}
	if len(obs.Objects) != len(before.Objects)+3 {
		t.Fatalf("objects should be the scene's then the overlay's: %d vs %d+3", len(obs.Objects), len(before.Objects))
	}
	if obs.Objects[0].Tag != "player" || obs.Objects[len(before.Objects)].Tag != "resume" {
		t.Fatalf("scene objects come first, then the overlay's: %+v", obs.Objects)
	}
	b, _ = json.Marshal(obs)
	if !strings.Contains(string(b), `"overlay":"pause"`) {
		t.Fatalf("json should carry the overlay: %s", b)
	}
}
