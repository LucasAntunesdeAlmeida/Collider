package collider

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"testing"
	"testing/fstest"

	"github.com/hajimehoshi/ebiten/v2"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

func TestAlphaDefaultsOpaqueAndClamps(t *testing.T) {
	o := Rect(10, 10, Red)
	if o.opacity() != 1 {
		t.Fatalf("a new object should be fully opaque, got %f", o.opacity())
	}
	if o.Alpha(0.25).opacity() != 0.25 {
		t.Fatalf("Alpha should set opacity, got %f", o.opacity())
	}
	if o.Alpha(7).opacity() != 1 || o.Alpha(-1).opacity() != 0 {
		t.Fatal("Alpha should clamp to 0..1")
	}
}

func TestGeoMCoversTheBoxShiftedByTheView(t *testing.T) {
	o := Rect(20, 10, Red).At(100, 50)
	m := o.geoM(30, 40, 4, 2) // a 4x2 source stretched over the 20x10 box
	if x, y := m.Apply(0, 0); !near(x, 60) || !near(y, 5) {
		t.Fatalf("source top-left should land on the box top-left minus the view, got (%f, %f)", x, y)
	}
	if x, y := m.Apply(4, 2); !near(x, 80) || !near(y, 15) {
		t.Fatalf("source bottom-right should land on the box bottom-right, got (%f, %f)", x, y)
	}
}

func TestFlipXMirrorsAroundTheCenter(t *testing.T) {
	o := Rect(20, 10, Red).At(100, 50).FlipX(true)
	m := o.geoM(0, 0, 20, 10)
	if x, y := m.Apply(0, 0); !near(x, 110) || !near(y, 45) {
		t.Fatalf("flipped, the source's left edge should draw on the right, got (%f, %f)", x, y)
	}
}

func TestRotateTurnsClockwiseAroundTheCenter(t *testing.T) {
	o := Rect(20, 10, Red).At(100, 100).Rotate(math.Pi / 2)
	m := o.geoM(0, 0, 20, 10)
	if x, y := m.Apply(10, 5); !near(x, 100) || !near(y, 100) {
		t.Fatalf("rotation must keep the center in place, got (%f, %f)", x, y)
	}
	if x, y := m.Apply(0, 0); !near(x, 105) || !near(y, 90) {
		t.Fatalf("a quarter turn clockwise should move the top-left corner to (105, 90), got (%f, %f)", x, y)
	}
	if o.box().W != 20 || o.box().H != 10 {
		t.Fatal("rotation is drawing only: the collider stays the upright box")
	}
}

func TestColorScaleCombinesColorTintAndAlpha(t *testing.T) {
	o := Rect(10, 10, color.RGBA{R: 200, G: 100, B: 50, A: 255}).
		Tint(color.RGBA{R: 255, G: 127, B: 0, A: 255}).Alpha(0.5)
	cs := o.colorScale(o.fill)
	if !near(float64(cs.A()), 0.5) {
		t.Fatalf("alpha should scale the color, got A=%f", cs.A())
	}
	if !near(float64(cs.R()), 200.0/255*0.5) || cs.B() != 0 {
		t.Fatalf("tint should multiply the fill (premultiplied by alpha), got R=%f B=%f", cs.R(), cs.B())
	}
	if g := float64(cs.G()); math.Abs(g-100.0/255*127.0/255*0.5) > 1e-3 {
		t.Fatalf("green should be fill x tint x alpha, got %f", g)
	}
}

func TestFlashOverridesColorsForItsDuration(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")
	o := s.Add(Rect(10, 10, Red).Tint(Blue))

	o.Flash(color.White, 0.1)
	cs := o.colorScale(o.fill)
	if cs.R() != 1 || cs.G() != 1 || cs.B() != 1 {
		t.Fatalf("a flashing object should draw in the flash color only, got %v", cs.String())
	}
	for range 10 {
		s.update(dt)
	}
	cs = o.colorScale(o.fill)
	if cs.R() == 1 && cs.G() == 1 {
		t.Fatal("the flash should wear off after its duration")
	}
}

func TestRestartRestoresDrawingEffects(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")
	o := s.Add(Sprite("").Alpha(0.5).Rotate(1).FlipX(true).Tint(Red))
	s.activate()

	o.Alpha(0).Rotate(3).FlipX(false).Tint(nil)
	o.Flash(White, 5)
	s.restart()
	if o.opacity() != 0.5 || o.rotation != 1 || !o.flipX || o.tint != Red || o.flashLeft != 0 {
		t.Fatal("restart should restore setup effects and stop any flash")
	}
}

func TestDrawingEveryKindWithEffectsIsSafe(t *testing.T) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 4, 4))); err != nil {
		t.Fatal(err)
	}
	UseAssets(fstest.MapFS{"hero.png": &fstest.MapFile{Data: buf.Bytes()}})
	defer UseAssets(nil)

	g := New("test", 800, 600)
	s := g.Scene("play")
	s.Add(Rect(20, 10, Red).Rotate(0.5).Alpha(0.5).Tint(Blue))
	s.Add(Text("HI").FlipX(true).Rotate(1).Alpha(0.3))
	s.Add(Button("OK").Tint(Green).Alpha(0.8))
	s.Add(Rect(10, 10, Red).Alpha(0)) // invisible: skipped
	s.Add(Text(""))                   // empty: nothing to draw
	flashing := s.Add(Rect(10, 10, Red))
	flashing.Flash(White, 1)
	hit := s.Add(Sprite("hero.png").FlipX(true))
	hit.Flash(White, 1) // the silhouette path

	s.draw(ebiten.NewImage(800, 600))
}
