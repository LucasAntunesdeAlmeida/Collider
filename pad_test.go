package collider

import (
	"math"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// fakePad is one scripted controller.
type fakePad struct {
	standard bool
	held     map[PadButton]bool
	axes     map[PadAxis]float64
}

// fakePads stands in for the connected gamepads, the way fakeInput
// stands in for the mouse and touch screen.
func fakePads(t *testing.T) map[ebiten.GamepadID]*fakePad {
	t.Helper()
	pads := map[ebiten.GamepadID]*fakePad{}

	oldIDs, oldLayout := appendGamepadIDs, standardLayout
	oldPressed, oldAxis := standardButtonPressed, standardAxisValue
	t.Cleanup(func() {
		appendGamepadIDs, standardLayout = oldIDs, oldLayout
		standardButtonPressed, standardAxisValue = oldPressed, oldAxis
	})

	appendGamepadIDs = func(ids []ebiten.GamepadID) []ebiten.GamepadID {
		for id := range pads {
			ids = append(ids, id)
		}
		slices.Sort(ids)
		return ids
	}
	standardLayout = func(id ebiten.GamepadID) bool { return pads[id] != nil && pads[id].standard }
	// Like Ebitengine, a pad without a standard mapping reads nothing
	// through the standard calls.
	standardButtonPressed = func(id ebiten.GamepadID, b PadButton) bool {
		p := pads[id]
		return p != nil && p.standard && p.held[b]
	}
	standardAxisValue = func(id ebiten.GamepadID, a PadAxis) float64 {
		if p := pads[id]; p != nil && p.standard {
			return p.axes[a]
		}
		return 0
	}
	return pads
}

func newFakePad(standard bool) *fakePad {
	return &fakePad{standard: standard, held: map[PadButton]bool{}, axes: map[PadAxis]float64{}}
}

func TestPadDownOnAnyStandardPad(t *testing.T) {
	pads := fakePads(t)
	r := &realInput{}

	if r.padConnected() || r.padDown(PadA) || r.padAxis(PadLeftX) != 0 {
		t.Fatal("no pads should read as nothing connected, nothing held")
	}

	// A pad without a standard mapping has no known layout: ignored.
	odd := newFakePad(false)
	odd.held[PadA] = true
	pads[0] = odd
	if r.padConnected() || r.padDown(PadA) {
		t.Fatal("a pad without a standard layout should not count")
	}

	one, two := newFakePad(true), newFakePad(true)
	pads[1], pads[2] = one, two
	if !r.padConnected() {
		t.Fatal("a standard pad should be connected")
	}
	two.held[PadStart] = true
	if !r.padDown(PadStart) {
		t.Fatal("a button held on the second pad should be down")
	}
	if r.padDown(PadA) || r.padDown(PadB) {
		t.Fatal("buttons nobody holds should be up")
	}
	one.held[PadUp] = true
	if !r.padDown(PadUp) || !r.padDown(PadStart) {
		t.Fatal("each pad's buttons should count")
	}
}

func TestPadAxisTakesThePadPushedFurthest(t *testing.T) {
	pads := fakePads(t)
	r := &realInput{}
	idle, used := newFakePad(true), newFakePad(true)
	pads[3], pads[7] = idle, used

	idle.axes[PadLeftX] = 0.04 // a resting stick is rarely exactly 0
	used.axes[PadLeftX] = -0.9
	used.axes[PadRightY] = 0.5
	if v := r.padAxis(PadLeftX); v != -0.9 {
		t.Fatalf("the pad in use should answer, got %v", v)
	}
	if v := r.padAxis(PadRightY); v != 0.5 {
		t.Fatalf("right stick Y = %v, want 0.5", v)
	}
	used.axes[PadLeftX] = 0
	if v := r.padAxis(PadLeftX); v != 0.04 {
		t.Fatalf("axes are raw, without a dead zone: got %v, want 0.04", v)
	}
}

func TestMixedInputMergesPads(t *testing.T) {
	pads := fakePads(t)
	m := &mixedInput{agent: &agentInput{keys: map[Key]bool{}}}

	if m.padConnected() {
		t.Fatal("no real pad and an unused agent pad: not connected")
	}
	p := newFakePad(true)
	p.held[PadX] = true
	p.axes[PadLeftX] = 0.7
	pads[0] = p
	if !m.padConnected() || !m.padDown(PadX) {
		t.Fatal("the person's pad should count")
	}
	if v := m.padAxis(PadLeftX); v != 0.7 {
		t.Fatalf("with the agent's stick centered the real one answers, got %v", v)
	}
	m.agent.pad.inject(Action{Pad: []PadButton{PadY}, StickX: -0.3})
	if !m.padDown(PadY) || !m.padDown(PadX) {
		t.Fatal("agent and person buttons should both count")
	}
	if v := m.padAxis(PadLeftX); v != -0.3 {
		t.Fatalf("a pushed agent stick answers, got %v", v)
	}
}

func TestHeadlessPad(t *testing.T) {
	// Real hardware must never leak into headless play.
	pads := fakePads(t)
	hw := newFakePad(true)
	hw.held[PadA] = true
	hw.axes[PadLeftX] = 1
	pads[0] = hw

	g := New("pad-test", 800, 600)
	s := g.Scene("play")
	type frame struct {
		connected, a, start bool
		x, y, rx            float64
	}
	var frames []frame
	s.OnUpdate(func(float64) {
		frames = append(frames, frame{
			g.PadConnected(), g.PadDown(PadA), g.PadDown(PadStart),
			g.PadAxis(PadLeftX), g.PadAxis(PadLeftY), g.PadAxis(PadRightX),
		})
	})
	g.Headless("play")
	frames = nil

	g.Step(Action{Keys: []Key{Right}})
	g.Step(Action{Pad: []PadButton{PadA, PadStart}, StickX: 0.5, StickY: -2})
	g.Step(Action{StickX: math.NaN()})
	g.Step(Action{})
	want := []frame{
		{false, false, false, 0, 0, 0}, // no pad until the agent uses one
		{true, true, true, 0.5, -1, 0}, // stick clamped to -1..1
		{true, false, false, 0, 0, 0},  // NaN centers; still connected
		{true, false, false, 0, 0, 0},  // connected for the rest of the run
	}
	if !slices.Equal(frames, want) {
		t.Fatalf("pad per frame =\n%v\nwant\n%v", frames, want)
	}

	g.Headless("play") // a fresh run starts without a pad
	frames = nil
	g.Step(Action{})
	if frames[0].connected {
		t.Fatal("Headless should start with no pad connected")
	}
}

func TestActArgsPad(t *testing.T) {
	g := New("pad-test", 800, 600)
	a, frames, err := g.actArgs(map[string]any{
		"pad": []any{"A", "Start", "Left"}, "stickX": 0.25, "stickY": -1.0, "frames": 3.0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(a.Pad, []PadButton{PadA, PadStart, PadLeft}) || a.StickX != 0.25 || a.StickY != -1 || frames != 3 {
		t.Fatalf("act args = %+v, %d frames", a, frames)
	}

	_, _, err = g.actArgs(map[string]any{"pad": []any{"Jump"}})
	if err == nil || !strings.Contains(err.Error(), `"Jump"`) || !strings.Contains(err.Error(), "Start") {
		t.Fatalf("an unknown pad button should error with the valid names, got %v", err)
	}
	if _, _, err = g.actArgs(map[string]any{"pad": []any{1.0}}); err == nil {
		t.Fatal("a non-string pad button should error")
	}
}

func TestPadNames(t *testing.T) {
	// The help text lists exactly the names act accepts.
	listed := strings.Split(padNamesHelp, ", ")
	var names []string
	for n := range padButtonNames {
		names = append(names, n)
	}
	sort.Strings(listed)
	sort.Strings(names)
	if !slices.Equal(listed, names) {
		t.Fatalf("help lists %v, act accepts %v", listed, names)
	}
	seen := map[PadButton]bool{}
	for _, b := range padButtonNames {
		if seen[b] {
			t.Fatalf("button %v has two names", b)
		}
		seen[b] = true
	}
}
