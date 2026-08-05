package collider

import (
	"bytes"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/LucasAntunesdeAlmeida/collider/internal/assets"
)

// defaultTextSize is the em size for Text objects unless TextSize is set.
const defaultTextSize = 24

// Font sources are cached process-wide by path: font files are immutable
// data, and Text objects need them before any Game exists. The empty
// path is the embedded Go font.
var (
	fontMu      sync.Mutex
	fontSources = map[string]*text.GoTextFaceSource{}
)

func fontSource(path string) *text.GoTextFaceSource {
	fontMu.Lock()
	defer fontMu.Unlock()
	if src, ok := fontSources[path]; ok {
		return src
	}
	data := goregular.TTF
	if path != "" {
		b, err := assets.ReadFile(path)
		if err != nil {
			panic("collider: cannot open font " + path + ": " + err.Error())
		}
		data = b
	}
	src, err := text.NewGoTextFaceSource(bytes.NewReader(data))
	if err != nil {
		panic("collider: cannot parse font " + path + ": " + err.Error())
	}
	fontSources[path] = src
	return src
}

func (o *Object) fontFace() text.Face {
	size := o.textSize
	if size == 0 {
		size = defaultTextSize
	}
	return &text.GoTextFace{Source: fontSource(o.fontPath), Size: size}
}

// Text creates a text object. Text objects render on screen and can be
// clicked, but never take part in collisions: a score label overlapping
// the ball must not bounce it.
func Text(str string) *Object {
	o := &Object{isText: true, textStr: str}
	o.measureText()
	return o
}

// Font sets a custom font from a TTF/OTF file path. Chainable.
func (o *Object) Font(path string) *Object {
	o.fontPath = path
	o.measureText()
	return o
}

// TextSize sets the font size in pixels. Chainable.
func (o *Object) TextSize(px float64) *Object {
	o.textSize = px
	o.measureText()
	return o
}

// TextColor sets the text color (default white). Chainable.
func (o *Object) TextColor(c Color) *Object {
	o.textColor = c
	return o
}

// SetText changes the text at runtime and re-measures the bounds.
func (o *Object) SetText(str string) {
	o.textStr = str
	o.measureText()
}

func (o *Object) measureText() {
	if !o.isText {
		return
	}
	if o.sizeSet {
		return
	}
	w, h := text.Measure(o.textStr, o.fontFace(), 0)
	o.w, o.h = w, h
}

func (o *Object) drawText(screen *ebiten.Image) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(o.X-o.w/2, o.Y-o.h/2)
	if o.textColor != nil {
		op.ColorScale.ScaleWithColor(o.textColor)
	}
	text.Draw(screen, o.textStr, o.fontFace(), op)
}
