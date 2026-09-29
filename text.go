package collider

import (
	"bytes"
	"math"
	"slices"
	"sync"

	"github.com/go-text/typesetting/font"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/LucasAntunesdeAlmeida/collider/internal/assets"
)

// defaultTextSize is the em size for Text objects unless TextSize is set.
const defaultTextSize = 24

// Font sources are cached process-wide by path: font files are immutable
// data, and Text objects need them before any Game exists. The empty
// path is the embedded Go font. A source is read and parsed the first
// time a text needs it, so a large fallback font costs nothing until a
// glyph only it has shows up.
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

// checkFont fails fast on a font path that does not exist, without
// reading or parsing the file: fallback fonts load lazily, and a typo
// should not wait for the first Chinese string to surface.
func checkFont(path string) {
	if path == "" {
		return
	}
	fontMu.Lock()
	_, loaded := fontSources[path]
	fontMu.Unlock()
	if loaded {
		return
	}
	if err := assets.Stat(path); err != nil {
		panic("collider: cannot open font " + path + ": " + err.Error())
	}
}

// glyphMapper is the parsed font behind a GoTextFaceSource (its
// UnsafeInternal value), which answers whether the font has a glyph
// for a rune without shaping or drawing anything.
type glyphMapper interface {
	NominalGlyph(r rune) (font.GID, bool)
}

// hasGlyph reports whether src has a glyph for r. ok is false when the
// source does not expose its font (a future Ebitengine), in which case
// the caller falls back to letting MultiFace choose among every font.
func hasGlyph(src *text.GoTextFaceSource, r rune) (has, ok bool) {
	m, ok := src.UnsafeInternal().(glyphMapper)
	if !ok {
		return false, false
	}
	_, has = m.NominalGlyph(r)
	return has, true
}

// fontsFor picks which fonts of a fallback chain str needs, as a bit
// mask (bit i = paths[i]): the first font always, and for each rune the
// first font that has it, or the last one when none does (it draws its
// missing-glyph box, as MultiFace would). Only fonts consulted are
// loaded. Line breaks and other control characters draw nothing and
// pick no font.
func fontsFor(paths []string, str string) uint64 {
	mask := uint64(1)
	if len(paths) == 1 {
		return mask
	}
	all := uint64(1)<<len(paths) - 1
	primary := fontSource(paths[0])
	for _, r := range str {
		if r < ' ' || r == 0x7F {
			continue
		}
		has, ok := hasGlyph(primary, r)
		if !ok {
			return all
		}
		if has {
			continue
		}
		pick := len(paths) - 1
		for i := 1; i < len(paths)-1; i++ {
			has, ok := hasGlyph(fontSource(paths[i]), r)
			if !ok {
				return all
			}
			if has {
				pick = i
				break
			}
		}
		mask |= 1 << pick
		if mask == all {
			break
		}
	}
	return mask
}

// maxFonts caps a fallback chain so a chain's fonts fit a uint64 mask.
const maxFonts = 64

// textFace is the face that measures and draws the object's text: the
// first font alone when it has every glyph, else a MultiFace of the
// fonts the text uses, in chain order. It is rebuilt only when the
// fonts in use change.
func (o *Object) textFace() text.Face {
	if o.face != nil && o.faceText == o.textStr {
		return o.face
	}
	paths := o.fontPaths
	if len(paths) == 0 {
		paths = defaultFont
	}
	mask := fontsFor(paths, o.textStr)
	o.faceText = o.textStr
	if o.face != nil && mask == o.faceMask {
		return o.face
	}
	size := o.textSize
	if size == 0 {
		size = defaultTextSize
	}
	o.faceMask = mask
	if mask == 1 {
		o.face = &text.GoTextFace{Source: fontSource(paths[0]), Size: size}
		return o.face
	}
	faces := make([]text.Face, 0, len(paths))
	for i, p := range paths {
		if mask&(1<<i) != 0 {
			faces = append(faces, &text.GoTextFace{Source: fontSource(p), Size: size})
		}
	}
	mf, err := text.NewMultiFace(faces...)
	if err != nil { // only when directions disagree, which GoTextFaces never do
		panic("collider: cannot combine fonts: " + err.Error())
	}
	o.face = mf
	return mf
}

// defaultFont is the chain of a text with no Font: the embedded Go font.
var defaultFont = []string{""}

// Text creates a text object. Text objects render on screen and can be
// clicked, but never take part in collisions: a score label overlapping
// the ball must not bounce it.
func Text(str string) *Object {
	o := &Object{isText: true, visual: true, textStr: str}
	o.measureText()
	return o
}

// Font sets the text's font from TTF/OTF file paths, each read once and
// cached. With several, the first draws every glyph it has and the
// others, in order, fill in the rest: a pixel font for Latin and
// Cyrillic, then a CJK font for Chinese:
//
//	engine.Text(s).Font("fonts/pixel.ttf", "fonts/cjk.ttf")
//
// Every font draws at the same TextSize, and every glyph of a line sits
// on one baseline. A line is as tall as the largest ascent plus the
// largest descent among the fonts the text actually uses, so text the
// first font draws entirely measures and draws exactly as with that
// font alone: adding a fallback never moves it. A fallback font is
// only read and parsed the first time a text needs one of its glyphs.
// Pixel fonts stay sharp at whole multiples of their design size; an
// 8px Latin font and a 12px CJK font are both sharp at 24 and 48.
//
// The empty path is the built-in Go font, the default; Font() with no
// paths restores it. A path that does not exist panics here. Chainable.
func (o *Object) Font(paths ...string) *Object {
	if len(paths) > maxFonts {
		panic("collider: too many fallback fonts")
	}
	for _, p := range paths {
		checkFont(p)
	}
	o.fontPaths = slices.Clone(paths)
	o.face, o.lines = nil, nil
	o.measureText()
	return o
}

// TextSize sets the font size in pixels. Chainable.
func (o *Object) TextSize(px float64) *Object {
	o.textSize = px
	o.face, o.lines = nil, nil
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

// Wrap breaks the text into lines no wider than width pixels, at
// spaces in Latin and Cyrillic text and between any two CJK
// characters, never before closing punctuation (，。！？」 , . ! ?) or
// after opening punctuation (「（ ( ), and never inside a word unless the
// word alone is wider than the line (then between its letters). A
// "\n" always starts a new line; spaces at the end of a wrapped line
// are dropped.
//
// The text's box becomes a column exactly width wide (for clicks, and
// agents' w) and as tall as its lines, centered on the text's position
// like every object: with TextAlign(AlignLeft) the lines start at
// X - width/2 whatever they say. The object's text (and what agents
// read) stays the unwrapped string. For a fixed text area (a dialog
// box) whose first line stays put however many lines follow, give the
// text a Size too: its lines then start at the top of that box. 0
// turns wrapping off. Applies to buttons' labels too. Chainable, and
// callable any time.
func (o *Object) Wrap(width float64) *Object {
	o.wrapW = max(0, width)
	o.lines = nil
	o.measureText()
	return o
}

// Align is how the lines of a text line up in its box (TextAlign).
type Align int

const (
	AlignCenter Align = iota // each line centered: the default
	AlignLeft                // lines start at the box's left edge
	AlignRight               // lines end at the box's right edge
)

// TextAlign sets how the lines of a text line up in its box: centered
// (the default), left or right. The box is the widest line, the Wrap
// column or the Size, and stays centered on the text's position; a single
// line without Wrap fills its box, so it only matters for text with
// several lines, or with Wrap. Chainable, and callable any time.
func (o *Object) TextAlign(a Align) *Object {
	o.align = a
	return o
}

// LineHeight sets the distance in pixels from one line's baseline to
// the next, for text with several lines. The default (0) is the fonts'
// own line height: the largest ascent plus the largest descent (plus
// any line gap) of the fonts in use, so lines of a pixel font touch;
// 1.5 times the text size is a comfortable paragraph. The box runs from
// the first line's top to the last line's bottom. Chainable, and
// callable any time.
func (o *Object) LineHeight(px float64) *Object {
	o.lineH = max(0, px)
	o.lines = nil
	o.measureText()
	return o
}

// alignX is where a line w wide starts in a box boxW wide, on a whole
// pixel.
func alignX(a Align, boxW, w float64) float64 {
	switch a {
	case AlignLeft:
		return 0
	case AlignRight:
		return math.Round(boxW - w)
	}
	return math.Round((boxW - w) / 2)
}

// textLines is a text laid out in lines: each line's advance, the
// distance between baselines, and the size of the block (the widest
// line, or the Wrap column; the first line's top to the last line's
// bottom).
type textLines struct {
	lines  []string
	widths []float64
	pitch  float64
	w, h   float64
}

// textLayout lays out the object's text with its face, Wrap and
// LineHeight, reusing the last layout while none of them changed.
func (o *Object) textLayout() *textLines {
	face := o.textFace()
	if o.lines != nil && o.linesText == o.textStr && o.linesFace == face {
		return o.lines
	}
	l := &textLines{}
	o.lines, o.linesText, o.linesFace = l, o.textStr, face
	if o.textStr == "" {
		return l
	}
	advance := func(s string) float64 {
		w, _ := text.Measure(s, face, 0)
		return w
	}
	l.lines = wrapLines(o.textStr, o.wrapW, advance)
	l.widths = make([]float64, len(l.lines))
	for i, line := range l.lines {
		l.widths[i] = advance(line)
		l.w = max(l.w, l.widths[i])
	}
	m := face.Metrics()
	l.pitch = o.lineH
	if l.pitch == 0 {
		l.pitch = m.HAscent + m.HDescent + m.HLineGap
	}
	l.h = float64(len(l.lines)-1)*l.pitch + m.HAscent + m.HDescent
	if o.wrapW > 0 {
		l.w = max(l.w, o.wrapW)
	}
	return l
}

func (o *Object) measureText() {
	if o.sizeSet || (!o.isText && !o.isButton) {
		return
	}
	l := o.textLayout()
	w, h := l.w, l.h
	if o.isButton {
		w += buttonPadX * 2
		h += buttonPadY * 2
	}
	o.w, o.h = w, h
}

func (o *Object) drawText(screen *ebiten.Image, vx, vy float64) {
	o.drawLines(screen, o.textLayout(), o.w, o.geoM(vx, vy, o.w, o.h))
}

// drawLines draws laid-out lines in a box boxW wide whose top-left is
// at the origin of geo, each line placed by the alignment on whole
// pixels so pixel fonts stay sharp.
func (o *Object) drawLines(screen *ebiten.Image, l *textLines, boxW float64, geo ebiten.GeoM) {
	face := o.textFace()
	op := &text.DrawOptions{}
	op.ColorScale = o.colorScale(o.textColor)
	for i, line := range l.lines {
		if line == "" {
			continue
		}
		op.GeoM.Reset()
		op.GeoM.Translate(alignX(o.align, boxW, l.widths[i]), float64(i)*l.pitch)
		op.GeoM.Concat(geo)
		text.Draw(screen, line, face, op)
	}
}
