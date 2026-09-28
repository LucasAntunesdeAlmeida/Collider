package collider

import (
	"slices"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// fakeDevices scripts the hardware realInput reads, one frame at a
// time. Ebitengine answers for real devices only inside a running game
// loop, so a test drives these instead.
type fakeDevices struct {
	mouseX, mouseY int
	mouseClick     bool
	mouseHeld      bool
	unfocused      bool
	touches        []ebiten.TouchID
	justPressed    []ebiten.TouchID
	pos            map[ebiten.TouchID][2]int
}

func fakeInput(t *testing.T) *fakeDevices {
	t.Helper()
	d := &fakeDevices{pos: map[ebiten.TouchID][2]int{}}

	oldCursor, oldTouches := cursorPosition, appendTouchIDs
	oldPos, oldPressed, oldClick := touchPosition, appendJustPressedTouchIDs, mouseClickJustPressed
	oldHeld, oldFocus := mouseButtonPressed, windowFocused
	t.Cleanup(func() {
		cursorPosition, appendTouchIDs = oldCursor, oldTouches
		touchPosition, appendJustPressedTouchIDs, mouseClickJustPressed = oldPos, oldPressed, oldClick
		mouseButtonPressed, windowFocused = oldHeld, oldFocus
	})

	cursorPosition = func() (int, int) { return d.mouseX, d.mouseY }
	appendTouchIDs = func(ids []ebiten.TouchID) []ebiten.TouchID { return append(ids, d.touches...) }
	appendJustPressedTouchIDs = func(ids []ebiten.TouchID) []ebiten.TouchID {
		return append(ids, d.justPressed...)
	}
	touchPosition = func(id ebiten.TouchID) (int, int) {
		p := d.pos[id]
		return p[0], p[1]
	}
	mouseClickJustPressed = func() bool { return d.mouseClick }
	mouseButtonPressed = func() bool { return d.mouseHeld }
	windowFocused = func() bool { return !d.unfocused }
	return d
}

// press puts a finger down. It counts as just pressed until the next
// frame, the way inpututil reports it.
func (d *fakeDevices) press(id ebiten.TouchID, x, y int) {
	d.touches = append(d.touches, id)
	d.justPressed = append(d.justPressed, id)
	d.pos[id] = [2]int{x, y}
}

func (d *fakeDevices) lift(id ebiten.TouchID) {
	d.touches = slices.DeleteFunc(d.touches, func(t ebiten.TouchID) bool { return t == id })
	delete(d.pos, id)
}

// frame ends the tick: just-pressed state lives for one frame only.
func (d *fakeDevices) frame() {
	d.justPressed = d.justPressed[:0]
	d.mouseClick = false
}

func TestTapIsAClick(t *testing.T) {
	d := fakeInput(t)
	r := &realInput{}

	if r.clickJustPressed() {
		t.Fatal("no input should be no click")
	}

	d.press(1, 120, 340)
	if !r.clickJustPressed() {
		t.Fatal("a finger landing should read as a click")
	}
	if x, y := r.cursor(); x != 120 || y != 340 {
		t.Fatalf("cursor should follow the finger, got %v,%v", x, y)
	}

	d.frame()
	if r.clickJustPressed() {
		t.Fatal("a held finger should not click again")
	}
}

func TestSecondFingerClicksWhereItLanded(t *testing.T) {
	d := fakeInput(t)
	r := &realInput{}

	d.press(1, 100, 100)
	r.clickJustPressed()
	r.cursor()
	d.frame()

	// The first finger stays down while a second lands elsewhere. The
	// click belongs to the new finger, so the cursor has to be there.
	d.press(2, 700, 500)
	if !r.clickJustPressed() {
		t.Fatal("a second finger landing should read as a click")
	}
	if x, y := r.cursor(); x != 700 || y != 500 {
		t.Fatalf("click should report the finger that just landed, got %v,%v", x, y)
	}
	d.frame()

	// It lifts, and the finger still down takes the cursor back.
	d.lift(2)
	if x, y := r.cursor(); x != 100 || y != 100 {
		t.Fatalf("cursor should fall back to the finger still down, got %v,%v", x, y)
	}
}

func TestLiftedFingerLeavesTheCursor(t *testing.T) {
	d := fakeInput(t)
	r := &realInput{}

	d.press(1, 250, 250)
	r.cursor()
	d.frame()

	// A touch device reports no mouse position worth reading, so the
	// cursor stays where the finger left it.
	d.lift(1)
	if x, y := r.cursor(); x != 250 || y != 250 {
		t.Fatalf("a lifted finger should leave the cursor where it was, got %v,%v", x, y)
	}

	// A real mouse move takes it back.
	d.mouseX, d.mouseY = 10, 20
	if x, y := r.cursor(); x != 10 || y != 20 {
		t.Fatalf("a mouse move should take the cursor back, got %v,%v", x, y)
	}
}

func TestMouseIsUntouched(t *testing.T) {
	d := fakeInput(t)
	r := &realInput{}

	d.mouseX, d.mouseY = 400, 300
	if x, y := r.cursor(); x != 400 || y != 300 {
		t.Fatalf("cursor should read the mouse, got %v,%v", x, y)
	}
	if r.clickJustPressed() {
		t.Fatal("an idle mouse should not click")
	}

	d.mouseClick = true
	if !r.clickJustPressed() {
		t.Fatal("a mouse press should still click")
	}
}

func TestMixedInputReadsTouch(t *testing.T) {
	d := fakeInput(t)
	m := &mixedInput{agent: &agentInput{}}

	d.press(1, 42, 84)
	if !m.clickJustPressed() {
		t.Fatal("windowed agent play should still see a tap")
	}
	if x, y := m.cursor(); x != 42 || y != 84 {
		t.Fatalf("windowed agent play should follow the finger, got %v,%v", x, y)
	}

	// An agent click still wins: it is answering for its own pointer.
	m.agent.click, m.agent.x, m.agent.y = true, 9, 9
	if x, y := m.cursor(); x != 9 || y != 9 {
		t.Fatalf("agent input should take precedence, got %v,%v", x, y)
	}
}

func TestMouseDownReadsButtonAndFingers(t *testing.T) {
	d := fakeInput(t)
	r := &realInput{}

	if r.pointerDown() {
		t.Fatal("nothing held should not be down")
	}
	d.mouseHeld = true
	if !r.pointerDown() {
		t.Fatal("a held left button should be down")
	}
	d.mouseHeld = false

	d.press(1, 50, 60)
	d.frame()
	if !r.pointerDown() {
		t.Fatal("a finger resting on the glass should be down")
	}
	d.press(2, 70, 80)
	d.lift(1)
	if !r.pointerDown() {
		t.Fatal("any finger still down keeps the pointer down")
	}
	d.lift(2)
	if r.pointerDown() {
		t.Fatal("no fingers and no button should be up")
	}
}

func TestMixedInputOrsPointerDown(t *testing.T) {
	d := fakeInput(t)
	m := &mixedInput{agent: &agentInput{}}

	if m.pointerDown() {
		t.Fatal("idle should be up")
	}
	d.mouseHeld = true
	if !m.pointerDown() {
		t.Fatal("the person's held button should count")
	}
	d.mouseHeld = false
	m.agent.down, m.agent.x, m.agent.y = true, 30, 40
	if !m.pointerDown() {
		t.Fatal("the agent's held pointer should count")
	}
	d.mouseX, d.mouseY = 500, 500
	if x, y := m.cursor(); x != 30 || y != 40 {
		t.Fatalf("a held agent pointer answers for the cursor, got %v,%v", x, y)
	}
}

func TestFocused(t *testing.T) {
	d := fakeInput(t)
	r := &realInput{}
	if !r.focused() {
		t.Fatal("a focused window should report focus")
	}
	d.unfocused = true
	if r.focused() {
		t.Fatal("an unfocused window should report it")
	}
	// Agent play never loses focus: nobody to switch away, and a
	// person in their MCP client must not pause the agent's game.
	if !(&mixedInput{agent: &agentInput{}}).focused() || !(&agentInput{}).focused() {
		t.Fatal("agent input should always be focused")
	}
}

func TestHeadlessMouseDownAndFocus(t *testing.T) {
	g := New("down-test", 800, 600)
	s := g.Scene("play")
	var downs []bool
	var at [2]float64
	s.OnUpdate(func(float64) {
		downs = append(downs, g.MouseDown())
		if g.MouseDown() {
			at[0], at[1] = g.Mouse()
		}
		if !g.Focused() {
			t.Fatal("headless play is always focused")
		}
	})
	g.Headless("play")
	downs = nil

	g.Step(Action{})
	g.Step(Action{MouseDown: true, MouseX: 120, MouseY: 80})
	g.Step(Action{Click: true, MouseX: 10, MouseY: 20})
	g.Step(Action{})
	if want := []bool{false, true, true, false}; !slices.Equal(downs, want) {
		t.Fatalf("MouseDown per frame = %v, want %v (Click implies down)", downs, want)
	}
	if at != [2]float64{10, 20} {
		t.Fatalf("pointer position while down, got %v", at)
	}
}
