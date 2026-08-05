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

// keyNames maps the exported key constants to the names agents use in
// MCP calls and observations.
var keyNames = map[string]Key{
	"Left": Left, "Right": Right, "Up": Up, "Down": Down,
	"Space": Space, "Enter": Enter, "Esc": Esc,
	"W": W, "A": A, "S": S, "D": D, "F": F, "R": R,
}

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
