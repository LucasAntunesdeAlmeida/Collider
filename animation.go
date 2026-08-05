package collider

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// animation is one named sprite-sheet strip: a horizontal row of
// equally sized frames in a single image.
type animation struct {
	path   string
	frames int
	fps    float64
	imgs   []*ebiten.Image
}

// Animation defines a named animation from a horizontal sprite strip:
// the image is cut into `frames` equal slices played at `fps`.
// Chainable. Define several (run, jump, death...) and switch with Play.
func (o *Object) Animation(name, path string, frames int, fps float64) *Object {
	if o.anims == nil {
		o.anims = map[string]*animation{}
	}
	a := &animation{path: path, frames: frames, fps: fps}
	o.anims[name] = a
	if o.scene != nil {
		a.load(o.scene.game, o)
	}
	return o
}

// Play switches to a looping animation. Playing the animation that is
// already active does nothing, so calling it every frame is fine.
func (o *Object) Play(name string) {
	o.play(name, false)
}

// PlayOnce switches to an animation that runs once and holds its last
// frame: death, explosion, one-shot attacks.
func (o *Object) PlayOnce(name string) {
	o.play(name, true)
}

func (o *Object) play(name string, once bool) {
	if o.curAnim == name {
		return
	}
	if _, ok := o.anims[name]; !ok {
		panic("collider: unknown animation " + name)
	}
	o.curAnim = name
	o.animTime = 0
	o.animOnce = once
}

// load cuts the strip into frames; called when the object joins a scene
// (or immediately if it already has one).
func (a *animation) load(g *Game, o *Object) {
	if a.imgs != nil {
		return
	}
	sheet := g.assets.Image(a.path)
	b := sheet.Bounds()
	fw := b.Dx() / a.frames
	for i := range a.frames {
		r := image.Rect(b.Min.X+i*fw, b.Min.Y, b.Min.X+(i+1)*fw, b.Max.Y)
		a.imgs = append(a.imgs, sheet.SubImage(r).(*ebiten.Image))
	}
	if !o.sizeSet && o.w == 0 && o.h == 0 {
		o.w, o.h = float64(fw), float64(b.Dy())
	}
}

// animate advances the active animation; runs every frame from the
// scene update.
func (o *Object) animate(dt float64) {
	if o.curAnim == "" {
		return
	}
	o.animTime += dt
}

// animFrame returns the image to draw for the active animation, or nil.
func (o *Object) animFrame() *ebiten.Image {
	if o.curAnim == "" {
		return nil
	}
	a := o.anims[o.curAnim]
	if len(a.imgs) == 0 {
		return nil
	}
	return a.imgs[o.frameIndex(a)]
}

// frameIndex maps elapsed time to a frame: looping by default, clamped
// to the last frame for PlayOnce.
func (o *Object) frameIndex(a *animation) int {
	f := int(o.animTime * a.fps)
	if o.animOnce {
		if f >= a.frames {
			return a.frames - 1
		}
		return f
	}
	return f % a.frames
}
