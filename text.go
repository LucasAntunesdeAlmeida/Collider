package collider

import (
	"bytes"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

// defaultTextSize is the em size for Text objects. A custom size and
// font can come later if a real example demands them.
const defaultTextSize = 24

var (
	fontOnce sync.Once
	fontSrc  *text.GoTextFaceSource
)

func fontFace() text.Face {
	fontOnce.Do(func() {
		src, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
		if err != nil {
			panic("collider: cannot load embedded font: " + err.Error())
		}
		fontSrc = src
	})
	return &text.GoTextFace{Source: fontSrc, Size: defaultTextSize}
}

// Text creates a text object. Text objects render on screen and can be
// clicked, but never take part in collisions: a score label overlapping
// the ball must not bounce it.
func Text(str string) *Object {
	o := &Object{isText: true, textStr: str}
	o.measureText()
	return o
}

// SetText changes the text at runtime and re-measures the bounds.
func (o *Object) SetText(str string) {
	o.textStr = str
	if !o.sizeSet {
		o.measureText()
	}
}

func (o *Object) measureText() {
	w, h := text.Measure(o.textStr, fontFace(), 0)
	o.w, o.h = w, h
}

func (o *Object) drawText(screen *ebiten.Image) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(o.X-o.w/2, o.Y-o.h/2)
	text.Draw(screen, o.textStr, fontFace(), op)
}
