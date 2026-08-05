package assets

import (
	"bytes"
	"io"
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
// game that plays no audio never touches the sound device.
var (
	ctxOnce sync.Once
	ctx     *audio.Context
)

func context() *audio.Context {
	ctxOnce.Do(func() { ctx = audio.NewContext(SampleRate) })
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

// PlaySound plays a short effect, fire and forget. The decoded PCM is
// cached by path, so repeated plays are allocation-cheap.
func (c *Cache) PlaySound(path string) {
	c.mu.Lock()
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
	c.mu.Unlock()

	context().NewPlayerFromBytes(pcm).Play()
}

// PlayMusic starts a path looping forever and returns the player, so
// the caller can stop it on scene switches.
func (c *Cache) PlayMusic(path string) *audio.Player {
	c.mu.Lock()
	data := c.fileLocked(path)
	c.mu.Unlock()

	stream := decodeAudio(path, bytes.NewReader(data))
	loop := audio.NewInfiniteLoop(stream, stream.Length())
	p, err := context().NewPlayer(loop)
	if err != nil {
		panic("collider: cannot play music " + path + ": " + err.Error())
	}
	p.Play()
	return p
}
