package collider

import (
	"math"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// mono measures like a monospaced font: one unit per character, two
// for CJK (full width), none for combining marks.
func mono(s string) float64 {
	var w float64
	for _, r := range s {
		switch {
		case cjk(r):
			w += 2
		case r >= 0x300 && r <= 0x36F:
		default:
			w++
		}
	}
	return w
}

func TestWrapLines(t *testing.T) {
	cases := []struct {
		name  string
		in    string
		width float64
		want  []string
	}{
		{"fits", "hello world", 20, []string{"hello world"}},
		{"at spaces", "the cave hums tonight", 10, []string{"the cave", "hums", "tonight"}},
		{"exact fit", "abcde fghij", 5, []string{"abcde", "fghij"}},
		{"spaces dropped at the break", "one   two", 4, []string{"one", "two"}},
		{"trailing spaces dropped", "one two   ", 7, []string{"one two"}},
		{"leading indent kept", "  one two", 7, []string{"  one", "two"}},
		{"newline always breaks", "a\nb c", 10, []string{"a", "b c"}},
		{"empty lines kept", "a\n\nb", 10, []string{"a", "", "b"}},
		{"crlf", "a\r\nb", 10, []string{"a", "b"}},
		{"empty", "", 10, []string{""}},
		{"long word by letters", "abcdefghij", 4, []string{"abcd", "efgh", "ij"}},
		{"long word after short", "go abcdefgh", 4, []string{"go", "abcd", "efgh"}},
		{"never splits a word that fits", "aa bbbb", 5, []string{"aa", "bbbb"}},
		{"cyrillic", "Пещера гудит сегодня", 12, []string{"Пещера гудит", "сегодня"}},
		{"accents stay on their letters", "café café", 4, []string{"café", "café"}},
		{"hyphen", "sci-fi story", 5, []string{"sci-", "fi", "story"}},
		{"leading minus is not a hyphen", "a -5", 3, []string{"a", "-5"}},
		{"nbsp never breaks", "100 km away", 6, []string{"100 km", "away"}},
		{"french spacing", "Bonjour ! Oui ?", 9, []string{"Bonjour !", "Oui ?"}},
		{"french guillemets", "« Salut » dit-il", 9, []string{"« Salut »", "dit-il"}},
		{"cjk between any two", "今晚洞穴在低鸣", 6, []string{"今晚洞", "穴在低", "鸣"}},
		{"kinsoku closing", "你好。世界", 4, []string{"你", "好。", "世界"}},
		{"kinsoku comma", "旅人，你好", 6, []string{"旅人，", "你好"}},
		{"kinsoku opening", "他说「你好」", 6, []string{"他说", "「你", "好」"}},
		{"kinsoku question", "什么？好", 6, []string{"什么？", "好"}},
		{"ellipsis never starts a line", "我……那", 4, []string{"我……", "那"}},
		{"cjk and latin mix", "第3波HP", 5, []string{"第3波", "HP"}},
		{"latin punctuation stays", "你好!", 4, []string{"你", "好!"}},
		{"no width: split only", "a b\nc", 0, []string{"a b", "c"}},
	}
	for _, c := range cases {
		got := wrapLines(c.in, c.width, mono)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: wrapLines(%q, %v) = %q, want %q", c.name, c.in, c.width, got, c.want)
		}
	}
}

func TestWrapLinesNeverExceedsWidth(t *testing.T) {
	text := "The horde grows at night; 夜里尸群越来越多，一波接一波。Орда растёт ночью. " +
		"Die Horde wächst in der Nacht, unaufhaltsam-gnadenlos. supercalifragilistic"
	for width := 1.0; width <= 30; width++ {
		lines := wrapLines(text, width, mono)
		joined := strings.Join(lines, "")
		for _, l := range lines {
			// A line may only be wider when it is a single character.
			if mono(l) > width && utf8.RuneCountInString(l) > 1 {
				t.Fatalf("width %v: line %q is %v wide", width, l, mono(l))
			}
			if strings.HasSuffix(l, " ") {
				t.Fatalf("width %v: line %q ends in a space", width, l)
			}
		}
		// Nothing is lost but the spaces at the breaks.
		strip := func(s string) string { return strings.ReplaceAll(s, " ", "") }
		if strip(joined) != strip(text) {
			t.Fatalf("width %v: lost text:\n%q\n%q", width, joined, text)
		}
	}
}

func TestWrappedTextBox(t *testing.T) {
	o := Text("the cave hums tonight").Wrap(80).TextSize(12)
	face := o.textFace()
	m := face.Metrics()
	l := o.textLayout()
	if len(l.lines) < 2 {
		t.Fatalf("expected several lines, got %q", l.lines)
	}
	for _, w := range l.widths {
		if w > 80 {
			t.Fatalf("a line is %v wide, over the 80 column", w)
		}
	}
	pitch := m.HAscent + m.HDescent + m.HLineGap
	wantH := float64(len(l.lines)-1)*pitch + m.HAscent + m.HDescent
	if o.w != 80 || o.h != wantH {
		t.Fatalf("box %vx%v, want the 80 column by %v", o.w, o.h, wantH)
	}

	// LineHeight spaces the baselines; the box follows.
	o.LineHeight(20)
	wantH = float64(len(l.lines)-1)*20 + m.HAscent + m.HDescent
	if o.h != wantH {
		t.Fatalf("with LineHeight(20) the box is %v tall, want %v", o.h, wantH)
	}

	// Unwrapping restores the one-line measure.
	o.Wrap(0)
	w, h := text.Measure("the cave hums tonight", face, 0)
	if o.w != w || o.h != h {
		t.Fatalf("unwrapped box %vx%v, want %vx%v", o.w, o.h, w, h)
	}
	if Text("").Wrap(100).w != 0 {
		t.Fatal("empty text has no box, wrapped or not")
	}
}

func TestNewlinesMakeLines(t *testing.T) {
	// Before Wrap existed, "\n" drew every line on top of the first.
	one := Text("SCORE").TextSize(12)
	two := Text("SCORE\nBEST").TextSize(12)
	m := two.textFace().Metrics()
	if two.h != one.h+m.HAscent+m.HDescent+m.HLineGap {
		t.Fatalf("two lines are %v tall, one line %v", two.h, one.h)
	}
	if two.w != one.w {
		t.Fatalf("two lines are as wide as the widest (%v), got %v", one.w, two.w)
	}
}

func TestTextAlignPlacesLines(t *testing.T) {
	// A line 30 wide in a 100 wide box, on whole pixels.
	if x := alignX(AlignLeft, 100, 30); x != 0 {
		t.Fatalf("left: %v", x)
	}
	if x := alignX(AlignRight, 100, 30); x != 70 {
		t.Fatalf("right: %v", x)
	}
	if x := alignX(AlignCenter, 100, 30); x != 35 {
		t.Fatalf("center: %v", x)
	}
	if x := alignX(AlignCenter, 100, 31); x != math.Round(34.5) {
		t.Fatalf("center on a whole pixel: %v", x)
	}
	o := Text("x")
	if o.align != AlignCenter {
		t.Fatal("centered is the default")
	}
	if o.TextAlign(AlignRight).align != AlignRight {
		t.Fatal("TextAlign should set the alignment")
	}
}

func TestWrapHeadless(t *testing.T) {
	g := New("wrap-test", 800, 600)
	s := g.Scene("play")
	long := "a paragraph long enough to need three or four lines here"
	p := s.Add(Text(long).Wrap(200).TextAlign(AlignLeft).At(300, 200))
	clicked := false
	p.OnClick(func() { clicked = true })
	b := s.Add(Button("A LONG BUTTON LABEL").TextSize(16).Wrap(150).At(400, 450))
	g.Headless("play")

	obs := g.Step(Action{})
	var got *ObjectObs
	for i := range obs.Objects {
		if obs.Objects[i].Text == long {
			got = &obs.Objects[i]
		}
	}
	if got == nil {
		t.Fatal("agents read the unwrapped text")
	}
	if got.W != 200 || got.H != p.h || p.h <= p.textLayout().pitch {
		t.Fatalf("observed %vx%v, want the 200 column by several lines", got.W, got.H)
	}

	// The column takes clicks where it is, even beside a short line.
	g.Step(Action{MouseX: 300 + 95, MouseY: 200 + p.h/2 - 2, Click: true})
	if !clicked {
		t.Fatal("a click inside the column should hit the text")
	}
	if b.w != 150+buttonPadX*2 || len(b.textLayout().lines) < 2 {
		t.Fatalf("a wrapped button is the column plus padding: %v, %q", b.w, b.textLayout().lines)
	}

	// SetText re-wraps; Restart rolls the text back and the lines follow.
	p.SetText("short")
	g.Step(Action{})
	if len(p.textLayout().lines) != 1 || p.w != 200 {
		t.Fatalf("short text: %q, box %v", p.textLayout().lines, p.w)
	}
	g.Restart("play")
	g.Step(Action{})
	if p.textStr != long || len(p.textLayout().lines) < 3 {
		t.Fatalf("after restart: %q in %d lines", p.textStr, len(p.textLayout().lines))
	}
}

func TestWrapCJKWithFallback(t *testing.T) {
	// The fallback font's glyphs wrap between any two of them.
	o := Text("你好世界你好世界").Font("", fallbackFont).TextSize(12).Wrap(40)
	l := o.textLayout()
	if len(l.lines) < 2 {
		t.Fatalf("CJK text should wrap, got %q", l.lines)
	}
	for _, w := range l.widths {
		if w > 40 {
			t.Fatalf("line %v wide over a 40 column: %q", w, l.lines)
		}
	}
	m := o.textFace().Metrics() // the combined face sets the pitch
	if l.pitch != m.HAscent+m.HDescent+m.HLineGap {
		t.Fatalf("pitch %v, want the fonts' line height", l.pitch)
	}
}

func TestWrapInsideAFixedArea(t *testing.T) {
	// A dialog box: Size fixes the area, Wrap lays the lines out in it.
	o := Text("the cave hums tonight").TextSize(12).Wrap(80).Size(80, 100)
	if o.w != 80 || o.h != 100 || len(o.textLayout().lines) < 2 {
		t.Fatalf("box %vx%v with lines %q", o.w, o.h, o.textLayout().lines)
	}
	o.SetText("x")
	if o.w != 80 || o.h != 100 {
		t.Fatalf("the area keeps its size as the text changes: %vx%v", o.w, o.h)
	}
}
