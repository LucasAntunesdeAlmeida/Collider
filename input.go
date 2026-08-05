package collider

import (
	"github.com/hajimehoshi/ebiten/v2"
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

// Key reports whether a key is currently held down.
func (g *Game) Key(k Key) bool {
	return ebiten.IsKeyPressed(k)
}

// Mouse returns the cursor position in game coordinates.
func (g *Game) Mouse() (x, y float64) {
	mx, my := ebiten.CursorPosition()
	return float64(mx), float64(my)
}
