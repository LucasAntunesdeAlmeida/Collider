package collider

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// fallbackFont is a tiny OFL subset of a CJK pixel font with only the
// glyphs 你好世界。 (testdata/OFL.txt); the built-in Go font has none of
// them, and the fallback has no Latin letters.
const fallbackFont = "testdata/fallback.ttf"

// freshFallback copies the fallback font to a path no other test has
// loaded, so the process-wide cache says whether this test loaded it.
func freshFallback(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(fallbackFont)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "fallback.ttf")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func loaded(path string) bool {
	fontMu.Lock()
	defer fontMu.Unlock()
	_, ok := fontSources[path]
	return ok
}

// measure is what a single font says about s at size px.
func measure(t *testing.T, path, s string, px float64) (w, h float64) {
	t.Helper()
	return text.Measure(s, &text.GoTextFace{Source: fontSource(path), Size: px}, 0)
}

func TestGlyphLookupReachesTheFont(t *testing.T) {
	// The lazy fallback relies on reading glyph coverage from the font
	// behind Ebitengine's face source; fail loudly if an upgrade hides it.
	if has, ok := hasGlyph(fontSource(""), 'A'); !ok || !has {
		t.Fatalf("hasGlyph('A') = %v, %v; want true, true", has, ok)
	}
	if has, ok := hasGlyph(fontSource(""), '你'); !ok || has {
		t.Fatalf("the Go font should not have 你: %v, %v", has, ok)
	}
	if has, _ := hasGlyph(fontSource(fallbackFont), '你'); !has {
		t.Fatal("the fallback font should have 你")
	}
}

func TestSinglePathFontUnchanged(t *testing.T) {
	o := Text("SCORE 12").Font(fallbackFont).TextSize(12)
	w, h := measure(t, fallbackFont, "SCORE 12", 12)
	if o.w != w || o.h != h {
		t.Fatalf("single font measures %vx%v, want %vx%v", o.w, o.h, w, h)
	}
	d := Text("SCORE 12").TextSize(12)
	w, h = measure(t, "", "SCORE 12", 12)
	if d.w != w || d.h != h {
		t.Fatalf("default font measures %vx%v, want %vx%v", d.w, d.h, w, h)
	}
	if _, ok := d.textFace().(*text.GoTextFace); !ok {
		t.Fatal("a single font draws with its own face, not a combined one")
	}
}

func TestFallbackFillsMissingGlyphs(t *testing.T) {
	const px = 24
	cjk := Text("你好").Font("", fallbackFont).TextSize(px)
	w, h := measure(t, fallbackFont, "你好", px)
	if cjk.w != w || cjk.h != h {
		t.Fatalf("你好 measures %vx%v, want the fallback's %vx%v", cjk.w, cjk.h, w, h)
	}

	// Drawing uses the same face, and every glyph is a real one from
	// the fallback: none is the first font's missing-glyph box (GID 0).
	for _, g := range text.AppendGlyphs(nil, "你好", cjk.textFace(), nil) {
		if g.GID == 0 {
			t.Fatalf("glyph at byte %d is the missing-glyph box", g.StartIndexInBytes)
		}
	}
	missing := 0
	for _, g := range text.AppendGlyphs(nil, "你好", &text.GoTextFace{Source: fontSource(""), Size: px}, nil) {
		if g.GID == 0 {
			missing++
		}
	}
	if missing != 2 {
		t.Fatalf("the Go font alone should draw 2 missing-glyph boxes, drew %d", missing)
	}
}

func TestMixedLineSharesOneBaseline(t *testing.T) {
	const px = 24
	o := Text("Hi 你好").Font("", fallbackFont).TextSize(px)
	latin := text.GoTextFace{Source: fontSource(""), Size: px}
	han := text.GoTextFace{Source: fontSource(fallbackFont), Size: px}
	lw, _ := text.Measure("Hi ", &latin, 0)
	hw, _ := text.Measure("你好", &han, 0)
	if o.w != lw+hw {
		t.Fatalf("mixed width %v, want %v + %v", o.w, lw, hw)
	}
	lm, hm := latin.Metrics(), han.Metrics()
	want := max(lm.HAscent, hm.HAscent) + max(lm.HDescent, hm.HDescent)
	if o.h != want {
		t.Fatalf("mixed height %v, want the largest ascent plus the largest descent %v", o.h, want)
	}
}

func TestFallbackNeverMovesLatinText(t *testing.T) {
	// Adding a fallback font leaves text the first font draws alone
	// exactly as it was: same size, same face.
	plain := Text("Hello, night").TextSize(16)
	with := Text("Hello, night").Font("", fallbackFont).TextSize(16)
	if plain.w != with.w || plain.h != with.h {
		t.Fatalf("with a fallback %vx%v, without %vx%v", with.w, with.h, plain.w, plain.h)
	}
	if _, ok := with.textFace().(*text.GoTextFace); !ok {
		t.Fatal("Latin-only text should draw with the first font alone")
	}
}

func TestFallbackLoadsLazily(t *testing.T) {
	path := freshFallback(t)
	o := Text("LEVEL UP").Font("", path).TextSize(12)
	if loaded(path) {
		t.Fatal("the fallback was parsed although no glyph needed it")
	}
	o.SetText("LEVEL 你好")
	if !loaded(path) {
		t.Fatal("the fallback should load when a glyph needs it")
	}
	if _, ok := o.textFace().(*text.MultiFace); !ok {
		t.Fatal("mixed text should draw with the combined face")
	}
	face := o.textFace()
	o.SetText("LEVEL 世界")
	if o.textFace() != face {
		t.Fatal("the face should be reused while the same fonts are in use")
	}
	o.SetText("LEVEL 2")
	if _, ok := o.textFace().(*text.GoTextFace); !ok {
		t.Fatal("back to Latin, the first font draws alone again")
	}
}

func TestFallbackOrderAndMissingEverywhere(t *testing.T) {
	// The first font that has a glyph draws it: here the fallback comes
	// first, so it draws 你 and the Go font the Latin letters it lacks.
	o := Text("A你").Font(fallbackFont, "").TextSize(12)
	aw, _ := measure(t, "", "A", 12)
	nw, _ := measure(t, fallbackFont, "你", 12)
	if o.w != aw+nw {
		t.Fatalf("width %v, want %v + %v", o.w, aw, nw)
	}
	// A glyph no font has is drawn (as a box) by the last font, like
	// MultiFace does; nothing panics.
	x := Text("☃").Font("", fallbackFont).TextSize(12)
	if x.w <= 0 {
		t.Fatal("a missing glyph still takes room")
	}
}

func TestFontPathsFailFast(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil || !strings.Contains(r.(string), "no-such-font.ttf") {
			t.Fatalf("a missing fallback font should panic at Font, got %v", r)
		}
	}()
	Text("hi").Font("", "fonts/no-such-font.ttf")
}

func TestFontResetsAndSizeRebuildsTheFace(t *testing.T) {
	o := Text("你").Font("", fallbackFont).TextSize(12)
	w12 := o.w
	o.TextSize(24)
	if o.w != 2*w12 {
		t.Fatalf("TextSize should re-measure with the new face: %v, want %v", o.w, 2*w12)
	}
	o.Font()
	if o.fontPaths != nil {
		t.Fatal("Font() restores the default font")
	}
	if _, ok := o.textFace().(*text.GoTextFace); !ok {
		t.Fatal("the default font draws alone")
	}
}

func TestFallbackHeadlessAndRestart(t *testing.T) {
	g := New("font-test", 800, 600)
	s := g.Scene("play")
	label := s.Add(Text("HP").Font("", fallbackFont).TextSize(24).At(400, 300))
	btn := s.Add(Button("好").Font("", fallbackFont).TextSize(24).At(400, 400))
	g.Headless("play")
	label.SetText("世界")
	obs := g.Step(Action{})
	w, h := measure(t, fallbackFont, "世界", 24)
	var saw bool
	for _, o := range obs.Objects {
		if o.Text == "世界" {
			saw = true
			if o.W != w || o.H != h {
				t.Fatalf("observed %vx%v, want %vx%v", o.W, o.H, w, h)
			}
		}
	}
	if !saw {
		t.Fatal("the label should be observed")
	}
	bw, bh := measure(t, fallbackFont, "好", 24)
	if btn.w != bw+buttonPadX*2 || btn.h != bh+buttonPadY*2 {
		t.Fatalf("button measures %vx%v with its label from the fallback", btn.w, btn.h)
	}

	// Restart rolls the text back; the face follows the text.
	g.Restart("play")
	g.Step(Action{})
	if label.textStr != "HP" {
		t.Fatalf("restart restored %q", label.textStr)
	}
	if _, ok := label.textFace().(*text.GoTextFace); !ok {
		t.Fatal("after restart the Latin text draws with the first font alone")
	}
}
