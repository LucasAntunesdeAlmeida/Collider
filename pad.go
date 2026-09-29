package collider

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// PadButton identifies a gamepad button in the standard layout (an Xbox
// style controller: A at the bottom of the face buttons). Nintendo and
// PlayStation pads map by position, so PadA is Cross on a PlayStation
// pad and B on a Nintendo one.
type PadButton = ebiten.StandardGamepadButton

const (
	PadA      = ebiten.StandardGamepadButtonRightBottom // bottom face button
	PadB      = ebiten.StandardGamepadButtonRightRight  // right face button
	PadX      = ebiten.StandardGamepadButtonRightLeft   // left face button
	PadY      = ebiten.StandardGamepadButtonRightTop    // top face button
	PadLB     = ebiten.StandardGamepadButtonFrontTopLeft
	PadRB     = ebiten.StandardGamepadButtonFrontTopRight
	PadLT     = ebiten.StandardGamepadButtonFrontBottomLeft  // left trigger, pulled
	PadRT     = ebiten.StandardGamepadButtonFrontBottomRight // right trigger, pulled
	PadBack   = ebiten.StandardGamepadButtonCenterLeft       // View / Select / Share
	PadStart  = ebiten.StandardGamepadButtonCenterRight      // Menu / Start / Options
	PadLStick = ebiten.StandardGamepadButtonLeftStick        // left stick, pressed in
	PadRStick = ebiten.StandardGamepadButtonRightStick       // right stick, pressed in
	PadUp     = ebiten.StandardGamepadButtonLeftTop          // d-pad
	PadDown   = ebiten.StandardGamepadButtonLeftBottom
	PadLeft   = ebiten.StandardGamepadButtonLeftLeft
	PadRight  = ebiten.StandardGamepadButtonLeftRight
)

// PadAxis identifies a stick axis in the standard layout.
type PadAxis = ebiten.StandardGamepadAxis

const (
	PadLeftX  = ebiten.StandardGamepadAxisLeftStickHorizontal  // -1 left .. 1 right
	PadLeftY  = ebiten.StandardGamepadAxisLeftStickVertical    // -1 up .. 1 down
	PadRightX = ebiten.StandardGamepadAxisRightStickHorizontal // -1 left .. 1 right
	PadRightY = ebiten.StandardGamepadAxisRightStickVertical   // -1 up .. 1 down
)

// padButtonNames maps the names agents send (MCP act "pad") to buttons:
// the constants without their Pad prefix.
var padButtonNames = map[string]PadButton{
	"A": PadA, "B": PadB, "X": PadX, "Y": PadY,
	"LB": PadLB, "RB": PadRB, "LT": PadLT, "RT": PadRT,
	"Back": PadBack, "Start": PadStart, "LStick": PadLStick, "RStick": PadRStick,
	"Up": PadUp, "Down": PadDown, "Left": PadLeft, "Right": PadRight,
}

const padNamesHelp = "A, B, X, Y, LB, RB, LT, RT, Back, Start, LStick, RStick, Up, Down, Left, Right"

// The gamepad reads realInput makes, swappable for tests like the
// pointer ones in input.go: a test has no game loop, so no pads.
var (
	appendGamepadIDs      = ebiten.AppendGamepadIDs
	standardLayout        = ebiten.IsStandardGamepadLayoutAvailable
	standardButtonPressed = ebiten.IsStandardGamepadButtonPressed
	standardAxisValue     = ebiten.StandardGamepadAxisValue
)

// standardPads lists the connected gamepads that have a standard layout
// mapping, reusing r.pads. Pads without one (rare: unknown adapters)
// are ignored, since their buttons have no known meaning.
func (r *realInput) standardPads() []ebiten.GamepadID {
	r.pads = appendGamepadIDs(r.pads[:0])
	n := 0
	for _, id := range r.pads {
		if standardLayout(id) {
			r.pads[n] = id
			n++
		}
	}
	r.pads = r.pads[:n]
	return r.pads
}

func (r *realInput) padDown(b PadButton) bool {
	for _, id := range r.standardPads() {
		if standardButtonPressed(id, b) {
			return true
		}
	}
	return false
}

// padAxis answers with the pad pushed furthest on this axis, so a
// second, idle controller never cancels the one in use.
func (r *realInput) padAxis(a PadAxis) float64 {
	v := 0.0
	for _, id := range r.standardPads() {
		if x := standardAxisValue(id, a); math.Abs(x) > math.Abs(v) {
			v = x
		}
	}
	return v
}

func (r *realInput) padConnected() bool { return len(r.standardPads()) > 0 }

// agentPad is an agent's virtual controller: the buttons held this
// frame, and the left stick.
type agentPad struct {
	buttons [ebiten.StandardGamepadButtonMax + 1]bool
	x, y    float64
	// on is set by the first action that uses the pad and stays set
	// for the run, so a game that switches its prompts (or pauses) on
	// PadConnected does not flicker as the agent lets go of the stick.
	on bool
}

func (a *agentInput) padDown(b PadButton) bool {
	return b >= 0 && int(b) < len(a.pad.buttons) && a.pad.buttons[b]
}

func (a *agentInput) padAxis(ax PadAxis) float64 {
	switch ax {
	case PadLeftX:
		return a.pad.x
	case PadLeftY:
		return a.pad.y
	}
	return 0
}

func (a *agentInput) padConnected() bool { return a.pad.on }

func (m *mixedInput) padDown(b PadButton) bool {
	return m.agent.padDown(b) || m.real.padDown(b)
}

// padAxis lets the agent's stick answer while it is pushed, else the
// real pads.
func (m *mixedInput) padAxis(a PadAxis) float64 {
	if v := m.agent.padAxis(a); v != 0 {
		return v
	}
	return m.real.padAxis(a)
}

func (m *mixedInput) padConnected() bool {
	return m.agent.padConnected() || m.real.padConnected()
}

// injectPad loads an action's pad input into the virtual controller.
func (p *agentPad) inject(a Action) {
	clear(p.buttons[:])
	for _, b := range a.Pad {
		if b >= 0 && int(b) < len(p.buttons) {
			p.buttons[b] = true
		}
	}
	p.x, p.y = clampStick(a.StickX), clampStick(a.StickY)
	if len(a.Pad) > 0 || p.x != 0 || p.y != 0 {
		p.on = true
	}
}

// clampStick keeps an agent's stick in -1..1 (NaN = centered).
func clampStick(v float64) float64 {
	if v != v {
		return 0
	}
	return min(1, max(-1, v))
}

// resolvePad turns agent-supplied pad button names into buttons,
// erroring on anything unknown.
func resolvePad(names []any) ([]PadButton, error) {
	var out []PadButton
	for _, n := range names {
		name, ok := n.(string)
		if !ok {
			return nil, fmt.Errorf("pad buttons must be strings, got %v", n)
		}
		b, ok := padButtonNames[name]
		if !ok {
			return nil, fmt.Errorf("unknown pad button %q: use one of %s", name, padNamesHelp)
		}
		out = append(out, b)
	}
	return out, nil
}

// PadDown reports whether a gamepad button is held on any connected
// controller with a standard layout (Xbox, PlayStation, Switch Pro,
// Steam Deck and most others). Agents hold pad buttons through
// Action.Pad.
func (g *Game) PadDown(b PadButton) bool {
	return g.input.padDown(b)
}

// PadAxis returns a stick axis from -1 to 1, raw: no dead zone, so
// apply your own (a resting stick rarely reads exactly 0; ignoring
// lengths under 0.2 is typical). Y grows downward like the screen.
// With several pads, the one pushed furthest on this axis answers.
// Agents set the left stick through Action.StickX and StickY.
func (g *Game) PadAxis(a PadAxis) float64 {
	return g.input.padAxis(a)
}

// PadConnected reports whether a gamepad with a standard layout is
// connected, so a game can show pad prompts ("A" instead of "ENTER").
// In headless and agent play it is false until an action first uses
// the pad, then true for the rest of the run.
func (g *Game) PadConnected() bool {
	return g.input.padConnected()
}
