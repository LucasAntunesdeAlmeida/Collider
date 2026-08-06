package collider

import (
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
}

// realInput reads the actual devices through Ebitengine.
type realInput struct{}

func (realInput) keyPressed(k Key) bool { return ebiten.IsKeyPressed(k) }

func (realInput) cursor() (float64, float64) {
	x, y := ebiten.CursorPosition()
	return float64(x), float64(y)
}

func (realInput) clickJustPressed() bool {
	return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
}

// agentInput is fed by Step calls instead of hardware.
type agentInput struct {
	keys  map[Key]bool
	x, y  float64
	click bool
}

// mixedInput merges the real keyboard and mouse with agent-injected
// input, for windowed MCP play: the person at the keyboard and the
// agent hold a controller each, wired to the same game.
type mixedInput struct {
	agent *agentInput
}

func (m *mixedInput) keyPressed(k Key) bool {
	return m.agent.keys[k] || ebiten.IsKeyPressed(k)
}

func (m *mixedInput) cursor() (float64, float64) {
	if m.agent.click {
		return m.agent.x, m.agent.y
	}
	x, y := ebiten.CursorPosition()
	return float64(x), float64(y)
}

func (m *mixedInput) clickJustPressed() bool {
	return m.agent.click || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
}

func (a *agentInput) keyPressed(k Key) bool      { return a.keys[k] }
func (a *agentInput) cursor() (float64, float64) { return a.x, a.y }
func (a *agentInput) clickJustPressed() bool     { return a.click }

// Key reports whether a key is currently held down.
func (g *Game) Key(k Key) bool {
	return g.input.keyPressed(k)
}

// Mouse returns the cursor position in game coordinates.
func (g *Game) Mouse() (x, y float64) {
	return g.input.cursor()
}
