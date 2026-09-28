package collider

import (
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Key identifies a keyboard key. The constants below cover common game
// controls; any ebiten.Key value also works.
type Key = ebiten.Key

const (
	Left  = ebiten.KeyArrowLeft
	Right = ebiten.KeyArrowRight
	Up    = ebiten.KeyArrowUp
	Down  = ebiten.KeyArrowDown
	Space = ebiten.KeySpace
	Enter = ebiten.KeyEnter
	Esc   = ebiten.KeyEscape
	W     = ebiten.KeyW
	A     = ebiten.KeyA
	S     = ebiten.KeyS
	D     = ebiten.KeyD
	F     = ebiten.KeyF
	R     = ebiten.KeyR
	P     = ebiten.KeyP
)

// keyNames maps every keyboard key's name to its key, so agents can
// press anything a player can. Names are ebiten's ("J", "Numpad1",
// "ShiftLeft", "ArrowLeft"), plus short aliases for the most common
// keys.
var keyNames = func() map[string]Key {
	m := map[string]Key{}
	for k := Key(0); k <= ebiten.KeyMax; k++ {
		if name := k.String(); name != "" {
			m[name] = k
		}
	}
	for alias, k := range map[string]Key{
		"Left": Left, "Right": Right, "Up": Up, "Down": Down,
		"Space": Space, "Enter": Enter, "Esc": Esc,
	} {
		m[alias] = k
	}
	return m
}()

// inputSource is where a Game reads its input from: the real keyboard
// and mouse in normal play, injected state in agent/headless play.
type inputSource interface {
	keyPressed(Key) bool
	cursor() (float64, float64)
	clickJustPressed() bool
	pointerDown() bool
	focused() bool
}

// The device reads realInput makes, kept as variables so tests can
// stand in for hardware: Ebitengine reports real devices only from
// inside a running game loop, and a test has no window.
var (
	cursorPosition            = ebiten.CursorPosition
	appendTouchIDs            = ebiten.AppendTouchIDs
	touchPosition             = ebiten.TouchPosition
	appendJustPressedTouchIDs = inpututil.AppendJustPressedTouchIDs
	mouseClickJustPressed     = func() bool {
		return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	}
	mouseButtonPressed = func() bool {
		return ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	}
	windowFocused = ebiten.IsFocused
)

// realInput reads the actual devices through Ebitengine.
//
// A finger is a mouse here. Ebitengine keeps touches on their own
// channel and never turns one into a click (in a browser it cancels
// the compatibility mouse events a tap would otherwise produce), so
// without this a phone or tablet could not press a single button.
type realInput struct {
	// Reused across frames so reading the touch state allocates
	// nothing in the main loop.
	touches, justPressed []ebiten.TouchID

	// The finger answering as the cursor. Several fingers can be on
	// the glass at once but only one can be the pointer, and the
	// engine reports them in no particular order, so the choice has to
	// be remembered rather than read off the list.
	pointer    ebiten.TouchID
	hasPointer bool

	// Where the last finger was, and whether the cursor is still
	// answering for it rather than for the mouse.
	touchX, touchY         float64
	fromTouch              bool
	lastMouseX, lastMouseY int
}

func (*realInput) keyPressed(k Key) bool { return ebiten.IsKeyPressed(k) }

// pointerTouch reports the finger acting as the cursor, if any. A
// finger that just landed takes the role, so the position read on a
// click frame belongs to the finger that caused the click; otherwise
// the finger already holding it keeps it until it lifts.
func (r *realInput) pointerTouch() (ebiten.TouchID, bool) {
	r.touches = appendTouchIDs(r.touches[:0])
	if len(r.touches) == 0 {
		r.hasPointer = false
		return 0, false
	}
	r.justPressed = appendJustPressedTouchIDs(r.justPressed[:0])
	switch {
	case len(r.justPressed) > 0:
		r.pointer = r.justPressed[0]
	case !r.hasPointer || !slices.Contains(r.touches, r.pointer):
		r.pointer = r.touches[0]
	}
	r.hasPointer = true
	return r.pointer, true
}

func (r *realInput) cursor() (float64, float64) {
	mx, my := cursorPosition()
	if id, ok := r.pointerTouch(); ok {
		x, y := touchPosition(id)
		r.touchX, r.touchY, r.fromTouch = float64(x), float64(y), true
		r.lastMouseX, r.lastMouseY = mx, my
		return r.touchX, r.touchY
	}
	// A finger that lifts leaves the cursor where it was. A touch
	// device has no mouse position worth reading, so snapping back to
	// it would fling whatever follows the cursor into a corner. A real
	// mouse move takes the cursor back.
	if r.fromTouch && mx == r.lastMouseX && my == r.lastMouseY {
		return r.touchX, r.touchY
	}
	r.fromTouch = false
	r.lastMouseX, r.lastMouseY = mx, my
	return float64(mx), float64(my)
}

func (r *realInput) clickJustPressed() bool {
	if mouseClickJustPressed() {
		return true
	}
	r.justPressed = appendJustPressedTouchIDs(r.justPressed[:0])
	return len(r.justPressed) > 0
}

// pointerDown is the left button held, or any finger on the glass.
func (r *realInput) pointerDown() bool {
	if mouseButtonPressed() {
		return true
	}
	r.touches = appendTouchIDs(r.touches[:0])
	return len(r.touches) > 0
}

func (*realInput) focused() bool { return windowFocused() }

// agentInput is fed by Step calls instead of hardware.
type agentInput struct {
	keys  map[Key]bool
	x, y  float64
	click bool
	down  bool // pointer held; a click implies it for its frame
}

// mixedInput merges the real keyboard and mouse with agent-injected
// input, for windowed MCP play: the person at the keyboard and the
// agent hold a controller each, wired to the same game.
type mixedInput struct {
	agent *agentInput
	real  realInput
}

func (m *mixedInput) keyPressed(k Key) bool {
	return m.agent.keys[k] || m.real.keyPressed(k)
}

func (m *mixedInput) cursor() (float64, float64) {
	if m.agent.click || m.agent.down {
		return m.agent.x, m.agent.y
	}
	return m.real.cursor()
}

func (m *mixedInput) clickJustPressed() bool {
	return m.agent.click || m.real.clickJustPressed()
}

func (m *mixedInput) pointerDown() bool {
	return m.agent.pointerDown() || m.real.pointerDown()
}

// focused is always true in windowed agent play: the person is often
// in another window (the MCP client) while the agent plays, and the
// game must not pause itself under the agent.
func (m *mixedInput) focused() bool { return true }

func (a *agentInput) keyPressed(k Key) bool      { return a.keys[k] }
func (a *agentInput) cursor() (float64, float64) { return a.x, a.y }
func (a *agentInput) clickJustPressed() bool     { return a.click }
func (a *agentInput) pointerDown() bool          { return a.down || a.click }
func (a *agentInput) focused() bool              { return true }

// Key reports whether a key is currently held down.
func (g *Game) Key(k Key) bool {
	return g.input.keyPressed(k)
}

// Mouse returns the cursor position in game coordinates.
func (g *Game) Mouse() (x, y float64) {
	return g.input.cursor()
}

// MouseDown reports whether the pointer is held: the left mouse button,
// or any finger on a touch screen. Drag, aim, charge, a virtual
// joystick: read it with Mouse every frame. Clicks (a press) still
// arrive through OnClick.
func (g *Game) MouseDown() bool {
	return g.input.pointerDown()
}

// Focused reports whether the game window has the keyboard focus, so
// a game can pause itself when the player switches away. Always true
// in headless and agent play (Step, MCP, Autopilot), where there is no
// player to switch away.
func (g *Game) Focused() bool {
	return g.input.focused()
}
