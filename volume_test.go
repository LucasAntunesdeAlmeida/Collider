package collider

import (
	"math"
	"testing"

	"github.com/LucasAntunesdeAlmeida/collider/internal/assets"
)

// Volume is tested without the audio device: the stored volumes, their
// clamping and how they combine, that a muted Sound does no work at
// all, and that headless games can call it all (and play sounds and
// music) safely. Headless runs never open the device.
func TestVolumeClampsAndStores(t *testing.T) {
	g := New("volume-test", 800, 600)
	setters := map[assets.Channel]func(float64){
		assets.Master: g.Volume,
		assets.Music:  g.MusicVolume,
		assets.Sounds: g.SoundVolume,
	}
	for ch, set := range setters {
		if v := g.assets.Volume(ch); v != 1 {
			t.Fatalf("channel %d: default volume should be 1, got %v", ch, v)
		}
		for _, c := range []struct{ in, want float64 }{
			{0.5, 0.5}, {-1, 0}, {2, 1}, {0, 0}, {1, 1}, {math.NaN(), 0}, {math.Inf(1), 1},
		} {
			set(c.in)
			if v := g.assets.Volume(ch); v != c.want {
				t.Fatalf("channel %d: set %v stored %v, want %v", ch, c.in, v, c.want)
			}
		}
	}
}

func TestVolumeChannelsMultiplyWithTheMaster(t *testing.T) {
	g := New("volume-test", 800, 600)
	g.Volume(0.5)
	g.MusicVolume(0.4)
	g.SoundVolume(0.8)
	if l := g.assets.Level(assets.Music); l != 0.2 {
		t.Fatalf("music level = %v, want 0.5*0.4", l)
	}
	if l := g.assets.Level(assets.Sounds); l != 0.4 {
		t.Fatalf("sound level = %v, want 0.5*0.8", l)
	}
	g.Volume(1)
	if m, s := g.assets.Level(assets.Music), g.assets.Level(assets.Sounds); m != 0.4 || s != 0.8 {
		t.Fatalf("channels keep their own volume under the master: %v, %v", m, s)
	}
	g.SoundVolume(0)
	if m, s := g.assets.Level(assets.Music), g.assets.Level(assets.Sounds); m != 0.4 || s != 0 {
		t.Fatalf("muting sounds leaves the music: %v, %v", m, s)
	}
}

func TestMutedSoundDoesNoWork(t *testing.T) {
	realRun(t)
	// A windowed game a player launched: Sound really plays here. The
	// file does not exist, so any work (reading, decoding, a player)
	// would panic.
	g := New("volume-test", 800, 600)
	g.SoundVolume(0)
	g.Sound("does-not-exist.wav")
	g.SoundVolume(1)
	g.Volume(0)
	g.Sound("does-not-exist.wav")

	// The control: with sound on, the same call does try to load it.
	g.Volume(1)
	defer func() {
		if recover() == nil {
			t.Fatal("an audible Sound should load its file")
		}
	}()
	g.Sound("does-not-exist.wav")
}

func TestVolumeIsSafeHeadless(t *testing.T) {
	g := New("volume-test", 800, 600)
	s := g.Scene("play")
	s.Music("does-not-exist.wav") // never loaded headless
	s.OnUpdate(func(float64) {
		g.Volume(0)
		g.MusicVolume(0.5)
		g.SoundVolume(0.5)
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
	if v := g.assets.Volume(assets.Master); v != 0.3 {
		t.Fatalf("volume should be stored headless too, got %v", v)
	}
	if l := g.assets.Level(assets.Music); l != 0.15 {
		t.Fatalf("music level should be stored headless too, got %v", l)
	}
}
