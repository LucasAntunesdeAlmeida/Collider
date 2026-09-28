package collider

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

// Alpha sets the object's opacity, from 0 (invisible) to 1 (opaque,
// the default): fades, ghosts, translucent auras. Chainable, and
// callable any time.
func (o *Object) Alpha(a float64) *Object {
	o.alpha = 1 - min(1, max(0, a))
	return o
}

// Tint multiplies the object's colors by c: one white sprite becomes a
// red, blue or pale variant, a scene goes cold under a blue tint. nil
// clears it. Chainable, and callable any time.
func (o *Object) Tint(c Color) *Object {
	o.tint = c
	return o
}

// Flash draws the object as a solid silhouette of color c for sec
// seconds, then back to normal: the classic hit flash is
// o.Flash(engine.White, 0.1). Opacity still applies; a new flash
// replaces a running one.
func (o *Object) Flash(c Color, sec float64) {
	o.flashColor = c
	o.flashLeft = sec
}

// Rotate turns the object to an absolute angle in radians, clockwise
// on screen, around its center: an arrow pointing where it flies.
// Drawing only: the collider stays the upright box. Applies to
// sprites, rectangles and text. Chainable, and callable any time.
func (o *Object) Rotate(rad float64) *Object {
	o.rotation = rad
	return o
}

// FlipX mirrors the object horizontally when on: a character facing
// the way it walks. Drawing only; applies to sprites, rectangles and
// text. Chainable, and callable any time.
func (o *Object) FlipX(on bool) *Object {
	o.flipX = on
	return o
}

// opacity is the object's alpha. It is stored inverted so the zero
// value of an Object is fully opaque.
func (o *Object) opacity() float64 {
	return 1 - o.alpha
}

// whitePixel is a 1x1 white image, cut from the middle of a 3x3 so its
// edges never sample transparency: scaled, rotated and colored, it
// draws every rectangle through the same path as sprites.
var whitePixel = func() *ebiten.Image {
	img := ebiten.NewImage(3, 3)
	img.Fill(color.White)
	return img.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
}()

// geoM places an iw x ih source so it covers the object's box, flipped
// and rotated around the center, shifted by the view origin (vx, vy).
func (o *Object) geoM(vx, vy, iw, ih float64) ebiten.GeoM {
	var m ebiten.GeoM
	m.Translate(-iw/2, -ih/2)
	sx, sy := o.w/iw, o.h/ih
	if o.flipX {
		sx = -sx
	}
	m.Scale(sx, sy)
	if o.rotation != 0 {
		m.Rotate(o.rotation)
	}
	m.Translate(o.X-vx, o.Y-vy)
	return m
}

// colorScale combines a base color (a fill, a text color, nil for a
// sprite's own pixels) with the tint and opacity. During a flash the
// flash color replaces both base and tint.
func (o *Object) colorScale(base Color) ebiten.ColorScale {
	var cs ebiten.ColorScale
	switch {
	case o.flashLeft > 0:
		cs.ScaleWithColor(o.flashColor)
	default:
		if base != nil {
			cs.ScaleWithColor(base)
		}
		if o.tint != nil {
			cs.ScaleWithColor(o.tint)
		}
	}
	cs.ScaleAlpha(float32(o.opacity()))
	return cs
}

// shaded returns base with the tint, flash and opacity applied, for
// the parts drawn as plain colored shapes (button plates).
func (o *Object) shaded(base Color) color.Color {
	cs := o.colorScale(base)
	return color.RGBA{
		R: uint8(min(1, cs.R()) * 0xFF), G: uint8(min(1, cs.G()) * 0xFF),
		B: uint8(min(1, cs.B()) * 0xFF), A: uint8(min(1, cs.A()) * 0xFF),
	}
}

// draw renders the object at its position minus the view origin
// (vx, vy): the camera shift, or zero for Fixed objects.
func (o *Object) draw(screen *ebiten.Image, vx, vy float64) {
	if o.alpha >= 1 || o.w == 0 || o.h == 0 {
		return // invisible, or nothing to draw
	}
	img := o.img
	if f := o.animFrame(); f != nil {
		img = f
	}
	switch {
	case o.isText:
		o.drawText(screen, vx, vy)
	case o.isButton:
		o.drawButton(screen, o.X-o.w/2-vx, o.Y-o.h/2-vy)
	case img != nil:
		b := img.Bounds()
		geo := o.geoM(vx, vy, float64(b.Dx()), float64(b.Dy()))
		if o.flashLeft > 0 {
			o.drawSilhouette(screen, img, geo)
			return
		}
		op := &ebiten.DrawImageOptions{GeoM: geo, ColorScale: o.colorScale(nil)}
		screen.DrawImage(img, op)
	case o.fill != nil:
		op := &ebiten.DrawImageOptions{GeoM: o.geoM(vx, vy, 1, 1), ColorScale: o.colorScale(o.fill)}
		screen.DrawImage(whitePixel, op)
	}
}

// drawSilhouette paints the sprite's shape in the flash color, keeping
// its transparency, so a hit reads instantly whatever the sprite's
// colors.
func (o *Object) drawSilhouette(screen, img *ebiten.Image, geo ebiten.GeoM) {
	r, g, b, _ := o.flashColor.RGBA()
	var cm colorm.ColorM
	cm.Scale(0, 0, 0, o.opacity())
	cm.Translate(float64(r)/0xFFFF, float64(g)/0xFFFF, float64(b)/0xFFFF, 0)
	colorm.DrawImage(screen, img, cm, &colorm.DrawImageOptions{GeoM: geo})
}
