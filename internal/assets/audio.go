package assets

import (
	"bytes"
	"io"
	"log"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

// SampleRate is the audio pipeline rate every stream is decoded to.
const SampleRate = 48000

// Ebitengine allows one audio context per process, so it is the only
// process-global state in the engine, created lazily on first use: a
// game that plays no audio never touches the sound device. ctx stays
// nil when there is no device that can play: the game runs silently.
var (
	ctxOnce sync.Once
	ctx     *audio.Context
)

// newContext opens the audio device, or reports why it cannot. A
// variable so tests can stand in for a machine without one.
var newContext = func(sampleRate int) (*audio.Context, error) {
	// Ebitengine opens the device in the background and turns a
	// failure into an error that ends the game loop, so make sure it
	// can open before creating the context.
	if err := probeDevice(sampleRate); err != nil {
		return nil, err
	}
	return audio.NewContext(sampleRate), nil
}

// logf reports the one audio problem worth a line: no device. A
// variable so tests can capture it.
var logf = log.Printf

// context returns the audio context, or nil when there is no working
// audio device, which it logs once.
func context() *audio.Context {
	ctxOnce.Do(func() {
		c, err := newContext(SampleRate)
		if err != nil {
			logf("collider: no audio device, playing silently: %v", err)
			return
		}
		ctx = c
	})
	return ctx
}

type audioStream interface {
	io.ReadSeeker
	Length() int64
}

func decodeAudio(path string, r io.ReadSeeker) audioStream {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".wav":
		s, err := wav.DecodeWithSampleRate(SampleRate, r)
		if err != nil {
			panic("collider: cannot decode sound " + path + ": " + err.Error())
		}
		return s
	case ".ogg":
		s, err := vorbis.DecodeWithSampleRate(SampleRate, r)
		if err != nil {
			panic("collider: cannot decode sound " + path + ": " + err.Error())
		}
		return s
	default:
		panic("collider: unsupported audio format (use .wav or .ogg): " + path)
	}
}

// PlaySound plays a short effect, fire and forget, at the Sounds level.
// The decoded PCM is cached by path, so repeated plays are
// allocation-cheap. Muted (a level of 0), it does no work at all: no
// decoding, no player. Without an audio device it still loads the
// file, so a missing or broken one fails the same everywhere, but
// plays nothing.
func (c *Cache) PlaySound(path string) {
	pcm, vol := c.sound(path)
	if vol == 0 {
		return
	}
	ac := context()
	if ac == nil {
		return
	}
	p := ac.NewPlayerFromBytes(pcm)
	p.SetVolume(vol)
	p.Play()
}

// sound returns a sound's decoded PCM and the level to play it at, or
// a level of 0 (and no work) when sounds are muted.
func (c *Cache) sound(path string) ([]byte, float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	vol := c.levelLocked(Sounds)
	if vol == 0 {
		return nil, 0
	}
	pcm, ok := c.sounds[path]
	if !ok {
		stream := decodeAudio(path, bytes.NewReader(c.fileLocked(path)))
		var err error
		pcm, err = io.ReadAll(stream)
		if err != nil {
			panic("collider: cannot read sound " + path + ": " + err.Error())
		}
		c.sounds[path] = pcm
	}
	return pcm, vol
}

// PlayMusic starts a path looping forever at the Music level and
// returns the player, so the caller can stop it on scene switches and
// change its volume. Muted, it still plays silently, so raising the
// volume brings the music back where it would be. Without an audio
// device it loads the file like PlaySound and returns nil.
func (c *Cache) PlayMusic(path string) *audio.Player {
	data, vol := c.music(path)
	stream := decodeAudio(path, bytes.NewReader(data))
	ac := context()
	if ac == nil {
		return nil
	}
	loop := audio.NewInfiniteLoop(stream, stream.Length())
	p, err := ac.NewPlayer(loop)
	if err != nil {
		panic("collider: cannot play music " + path + ": " + err.Error())
	}
	p.SetVolume(vol)
	p.Play()
	return p
}

// music returns a music file's bytes and the level to play it at.
func (c *Cache) music(path string) ([]byte, float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.fileLocked(path), c.levelLocked(Music)
}
