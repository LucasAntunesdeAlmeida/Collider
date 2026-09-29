package collider

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// buttonPadX and buttonPadY are the plate margins around the label when
// a button sizes itself.
const (
	buttonPadX = 28
	buttonPadY = 18
)

// Button creates a clickable plate with a centered label: menus, retry
// screens, anything a player presses. Like text, buttons never collide.
//
//	play := menu.Add(collider.Button("PLAY").At(400, 350))
//	play.OnClick(func() { g.Go("play") })
//
// Size follows the label unless Size is called; Color sets the plate
// color (the light and dark edges are derived from it) and TextColor
// the label.
func Button(label string) *Object {
	o := &Object{
		isButton:  true,
		visual:    true,
		textStr:   label,
		fill:      color.RGBA{R: 0x2A, G: 0x35, B: 0x50, A: 0xFF},
		textColor: White,
	}
	o.measureText()
	return o
}

// Color sets the object's color: the fill of a Rect, the plate of a
// Button. Chainable.
func (o *Object) Color(c Color) *Object {
	o.fill = c
	return o
}

// Visual marks an object as decoration: it is drawn but never takes
// part in collisions. Use it for backgrounds and scenery so they do not
// clutter the collision world. Chainable.
func (o *Object) Visual() *Object {
	o.visual = true
	return o
}

// Fixed pins the object to the screen: its position is in screen
// pixels, it ignores the scene camera, and clicks hit it where it is
// drawn. The HUD, a pause button, a virtual joystick. Like Visual
// objects, fixed ones never collide: they do not live in the world.
// Chainable.
func (o *Object) Fixed() *Object {
	o.fixed = true
	o.visual = true
	return o
}

// shade scales a color's channels, for deriving bevel edges.
func shade(c Color, factor float64) color.RGBA {
	r, g, b, _ := c.RGBA()
	clamp := func(v float64) uint8 {
		switch {
		case v < 0:
			return 0
		case v > 255:
			return 255
		}
		return uint8(v)
	}
	return color.RGBA{
		R: clamp(float64(r>>8) * factor),
		G: clamp(float64(g>>8) * factor),
		B: clamp(float64(b>>8) * factor),
		A: 0xFF,
	}
}

func (o *Object) drawButton(screen *ebiten.Image, left, top float64) {
	x := float32(left)
	y := float32(top)
	w := float32(o.w)
	h := float32(o.h)
	const edge = 4

	vector.FillRect(screen, x, y, w, h, o.shaded(o.fill), false)
	light := o.shaded(shade(o.fill, 2.1))
	dark := o.shaded(shade(o.fill, 0.45))
	vector.FillRect(screen, x, y, w, edge, light, false)
	vector.FillRect(screen, x, y, edge, h, light, false)
	vector.FillRect(screen, x, y+h-edge, w, edge, dark, false)
	vector.FillRect(screen, x+w-edge, y, edge, h, dark, false)

	face := o.textFace()
	tw, th := text.Measure(o.textStr, face, 0)
	op := &text.DrawOptions{}
	op.GeoM.Translate(left+o.w/2-tw/2, top+o.h/2-th/2)
	op.ColorScale = o.colorScale(o.textColor)
	text.Draw(screen, o.textStr, face, op)
}
