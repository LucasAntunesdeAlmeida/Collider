package collider

import (
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestKeyNamedDigits(t *testing.T) {
	digits := []Key{
		ebiten.KeyDigit0, ebiten.KeyDigit1, ebiten.KeyDigit2, ebiten.KeyDigit3, ebiten.KeyDigit4,
		ebiten.KeyDigit5, ebiten.KeyDigit6, ebiten.KeyDigit7, ebiten.KeyDigit8, ebiten.KeyDigit9,
	}
	for d, want := range digits {
		short := string(rune('0' + d))
		if got := KeyNamed(short); got != want {
			t.Errorf("KeyNamed(%q) = %v, want %v", short, got, want)
		}
		if got := KeyNamed("Digit" + short); got != want {
			t.Errorf("KeyNamed(%q) = %v, want %v", "Digit"+short, got, want)
		}
	}
	if KeyNamed("Numpad1") != ebiten.KeyNumpad1 {
		t.Error("the keypad keeps its own names")
	}
}

func TestKeyNamedLettersAndOthers(t *testing.T) {
	for name, want := range map[string]Key{
		"A": ebiten.KeyA, "J": ebiten.KeyJ, "Z": ebiten.KeyZ,
		"F1": ebiten.KeyF1, "Tab": ebiten.KeyTab, "Backspace": ebiten.KeyBackspace,
		"ShiftLeft": ebiten.KeyShiftLeft, "Comma": ebiten.KeyComma,
		"ArrowLeft": ebiten.KeyArrowLeft, "Escape": ebiten.KeyEscape,
		// The constants and their aliases agree.
		"W": W, "P": P, "M": M,
		"Left": Left, "Right": Right, "Up": Up, "Down": Down,
		"Space": Space, "Enter": Enter, "Esc": Esc,
	} {
		if got := KeyNamed(name); got != want {
			t.Errorf("KeyNamed(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestKeyNamedPanicsOnUnknownNames(t *testing.T) {
	for _, name := range []string{"", "Digit10", "10", "esc", "j", "Jump"} {
		func() {
			defer func() {
				r := recover()
				msg, _ := r.(string)
				if !strings.Contains(msg, "unknown key name") || !strings.Contains(msg, "\""+name+"\"") {
					t.Errorf("KeyNamed(%q) should panic naming the key, got %v", name, r)
				}
			}()
			KeyNamed(name)
		}()
	}
}

func TestKeyNamedDrivesInput(t *testing.T) {
	// The key a game looks up is the key agents press by the same name.
	g := New("test", 800, 600)
	s := g.Scene("play")
	one := KeyNamed("1")
	picks := 0
	s.OnUpdate(func(float64) {
		if g.Key(one) {
			picks++
		}
	})
	g.Headless("play")
	keys, err := g.resolveKeys([]any{"1"})
	if err != nil {
		t.Fatal(err)
	}
	g.Step(Action{Keys: keys})
	g.Step(Action{Keys: []Key{KeyNamed("Digit1")}})
	g.Step(Action{Keys: []Key{KeyNamed("2")}})
	if picks != 2 {
		t.Fatalf("pressing 1 by either name should register, 2 must not; got %d picks", picks)
	}
}
