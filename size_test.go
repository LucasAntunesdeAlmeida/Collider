package collider

import (
	"bytes"
	"image"
	"image/png"
	"testing"
	"testing/fstest"
)

// sizeIs reports whether o's measured size is exactly w x h.
func sizeIs(o *Object, w, h float64) bool {
	return o.Width() == w && o.Height() == h
}

func TestSizeOfShapesAndOverrides(t *testing.T) {
	r := Rect(30, 12, White)
	if !sizeIs(r, 30, 12) {
		t.Fatalf("Rect(30, 12) is %vx%v", r.Width(), r.Height())
	}
	r.Hitbox(4, 4)
	if !sizeIs(r, 30, 12) {
		t.Fatal("a Hitbox must not change the drawn size")
	}
	r.Size(50, 8)
	if !sizeIs(r, 50, 8) {
		t.Fatalf("Size(50, 8) reads back %vx%v", r.Width(), r.Height())
	}

	// A text with a Size is that box, whatever it says.
	area := Text("hello").Wrap(100).Size(120, 60)
	area.SetText("a much longer line that wraps into several lines")
	if !sizeIs(area, 120, 60) {
		t.Fatalf("a sized text area reads %vx%v, want its Size", area.Width(), area.Height())
	}
}

func TestSizeOfTextIsMeasuredAtOnce(t *testing.T) {
	// No game, no window, no frame: the size is there after each call.
	o := Text("HELLO")
	w, h := measure(t, "", "HELLO", defaultTextSize)
	if !sizeIs(o, w, h) || w == 0 {
		t.Fatalf("HELLO is %vx%v, want %vx%v", o.Width(), o.Height(), w, h)
	}

	// Cyrillic: twice the bytes of its letters, measured by glyph.
	const ru = "ПРИВЕТ"
	if has, _ := hasGlyph(fontSource(""), 'П'); !has {
		t.Fatal("the Go font should have Cyrillic")
	}
	o.SetText(ru)
	w, h = measure(t, "", ru, defaultTextSize)
	if !sizeIs(o, w, h) {
		t.Fatalf("%s is %vx%v, want %vx%v", ru, o.Width(), o.Height(), w, h)
	}
	if o.Width() >= float64(len(ru))*defaultTextSize/2 {
		t.Fatalf("%s is %v wide: measured by byte, not by glyph", ru, o.Width())
	}

	o.TextSize(48)
	w, h = measure(t, "", ru, 48)
	if !sizeIs(o, w, h) {
		t.Fatalf("TextSize(48): %vx%v, want %vx%v", o.Width(), o.Height(), w, h)
	}

	o.Font(fallbackFont)
	w, h = measure(t, fallbackFont, ru, 48)
	if !sizeIs(o, w, h) {
		t.Fatalf("Font(...): %vx%v, want %vx%v", o.Width(), o.Height(), w, h)
	}

	b := Button("OK")
	bw, bh := measure(t, "", "OK", defaultTextSize)
	if !sizeIs(b, bw+buttonPadX*2, bh+buttonPadY*2) {
		t.Fatalf("a button is its label plus the plate: %vx%v", b.Width(), b.Height())
	}
}

func TestSizeOfCJKThroughTheFallback(t *testing.T) {
	const px = 24
	o := Text("HI").Font("", fallbackFont).TextSize(px)
	latinW, latinH := measure(t, "", "HI", px)
	if !sizeIs(o, latinW, latinH) {
		t.Fatalf("Latin text ignores the fallback: %vx%v, want %vx%v", o.Width(), o.Height(), latinW, latinH)
	}

	// Every glyph comes from the fallback: exactly its measure.
	o.SetText("你好世界")
	w, h := measure(t, fallbackFont, "你好世界", px)
	if !sizeIs(o, w, h) || w == 0 {
		t.Fatalf("你好世界 is %vx%v, want the fallback's %vx%v", o.Width(), o.Height(), w, h)
	}

	// Mixed: the sum of both fonts' advances, as tall as both lines.
	o.SetText("HI你好")
	cw, ch := measure(t, fallbackFont, "你好", px)
	if o.Width() != latinW+cw || o.Height() < max(latinH, ch) {
		t.Fatalf("HI你好 is %vx%v, want %v wide, at least %v tall",
			o.Width(), o.Height(), latinW+cw, max(latinH, ch))
	}
}

func TestSizeOfWrappedText(t *testing.T) {
	const px = 24
	o := Text("你好世界你好世界").Font("", fallbackFont).TextSize(px)
	one := o.Width()
	lineH := o.Height()
	glyph, _ := measure(t, fallbackFont, "你", px)

	// A column three glyphs wide: 8 characters take 3 lines.
	col := glyph*3 + 1
	o.Wrap(col)
	pitch := o.textLayout().pitch
	if len(o.textLayout().lines) != 3 {
		t.Fatalf("lines %q, want 3", o.textLayout().lines)
	}
	if !sizeIs(o, col, 2*pitch+lineH) {
		t.Fatalf("wrapped: %vx%v, want the %v column by %v", o.Width(), o.Height(), col, 2*pitch+lineH)
	}

	// LineHeight moves every baseline after the first.
	o.LineHeight(40)
	if !sizeIs(o, col, 2*40+lineH) {
		t.Fatalf("LineHeight(40): %vx%v, want %vx%v", o.Width(), o.Height(), col, 2*40+lineH)
	}

	// "\n" makes lines without Wrap: the widest line, all the lines.
	o.Wrap(0).LineHeight(0)
	if !sizeIs(o, one, lineH) {
		t.Fatalf("unwrapped: %vx%v, want %vx%v", o.Width(), o.Height(), one, lineH)
	}
	o.SetText("你好\n你好世界")
	if !sizeIs(o, glyph*4, pitch+lineH) {
		t.Fatalf("two lines: %vx%v, want %vx%v", o.Width(), o.Height(), glyph*4, pitch+lineH)
	}
}

func TestSizeHeadless(t *testing.T) {
	g := New("size-test", 800, 600)
	s := g.Scene("play")
	label := s.Add(Text("SCORE 0").At(400, 20))
	wall := s.Add(Rect(40, 10, White).At(400, 500))
	var inFrame [2]float64
	label.OnUpdate(func(float64) {
		label.SetText("SCORE 1000000")
		inFrame = [2]float64{label.Width(), label.Height()}
		// Anchor the label's left edge at x = 10 in the same frame.
		label.X = 10 + label.Width()/2
	})
	g.Headless("play")
	obs := g.Step(Action{})

	w, h := measure(t, "", "SCORE 1000000", defaultTextSize)
	if inFrame != [2]float64{w, h} {
		t.Fatalf("in the frame of SetText: %v, want %vx%v", inFrame, w, h)
	}
	for _, o := range obs.Objects {
		if o.Text == "SCORE 1000000" && (o.W != w || o.X-o.W/2 != 10) {
			t.Fatalf("observed %+v, want %v wide from x = 10", o, w)
		}
	}
	if !sizeIs(wall, 40, 10) {
		t.Fatalf("wall is %vx%v", wall.Width(), wall.Height())
	}

	// Restart rolls the text back, and the size with it, measured with
	// the font the text has now.
	label.TextSize(12)
	g.Restart("play")
	if label.textStr != "SCORE 0" {
		t.Fatalf("restart restored %q", label.textStr)
	}
	w, h = measure(t, "", "SCORE 0", 12)
	if !sizeIs(label, w, h) {
		t.Fatalf("after restart %vx%v, want %vx%v", label.Width(), label.Height(), w, h)
	}
}

func TestSizeOfASprite(t *testing.T) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 7, 9))); err != nil {
		t.Fatal(err)
	}
	UseAssets(fstest.MapFS{"sprites/s.png": &fstest.MapFile{Data: buf.Bytes()}})
	defer UseAssets(nil)

	g := New("size-test", 800, 600)
	s := g.Scene("play")
	o := Sprite("sprites/s.png")
	if !sizeIs(o, 0, 0) {
		t.Fatal("a sprite's image is loaded when it joins a scene")
	}
	s.Add(o)
	if !sizeIs(o, 7, 9) {
		t.Fatalf("added sprite is %vx%v, want its image's 7x9", o.Width(), o.Height())
	}
	if big := s.Add(Sprite("sprites/s.png").Size(14, 18)); !sizeIs(big, 14, 18) {
		t.Fatalf("a sized sprite is %vx%v", big.Width(), big.Height())
	}
}
