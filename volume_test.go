package collider

import (
	"math"
	"testing"
)

// Volume is tested without the audio device: the stored master volume
// and its clamping, and that headless games can call it (and play
// sounds and music) safely. Headless runs never open the device.
func TestVolumeClampsAndStores(t *testing.T) {
	g := New("volume-test", 800, 600)
	if v := g.assets.Volume(); v != 1 {
		t.Fatalf("default volume should be 1, got %v", v)
	}
	for _, c := range []struct{ in, want float64 }{
		{0.5, 0.5}, {-1, 0}, {2, 1}, {0, 0}, {1, 1}, {math.NaN(), 0}, {math.Inf(1), 1},
	} {
		g.Volume(c.in)
		if v := g.assets.Volume(); v != c.want {
			t.Fatalf("Volume(%v) stored %v, want %v", c.in, v, c.want)
		}
	}
}

func TestVolumeIsSafeHeadless(t *testing.T) {
	g := New("volume-test", 800, 600)
	s := g.Scene("play")
	s.Music("does-not-exist.wav") // never loaded headless
	s.OnUpdate(func(float64) {
		g.Volume(0)
		g.Sound("does-not-exist.wav")
		g.Volume(0.3)
	})
	g.Headless("play")
	for range 3 {
		g.Step(Action{})
	}
	if g.musicPlayer != nil {
		t.Fatal("headless play must not start music")
	}
	if v := g.assets.Volume(); v != 0.3 {
		t.Fatalf("volume should be stored headless too, got %v", v)
	}
}
