package assets

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// tinyWAV is a valid 16-bit stereo PCM file of a few silent frames.
func tinyWAV() []byte {
	const frames, rate = 64, SampleRate
	data := frames * 4
	b := make([]byte, 44+data)
	copy(b[0:], "RIFF")
	binary.LittleEndian.PutUint32(b[4:], uint32(36+data))
	copy(b[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(b[16:], 16)
	binary.LittleEndian.PutUint16(b[20:], 1) // PCM
	binary.LittleEndian.PutUint16(b[22:], 2) // stereo
	binary.LittleEndian.PutUint32(b[24:], rate)
	binary.LittleEndian.PutUint32(b[28:], rate*4)
	binary.LittleEndian.PutUint16(b[32:], 4)
	binary.LittleEndian.PutUint16(b[34:], 16)
	copy(b[36:], "data")
	binary.LittleEndian.PutUint32(b[40:], uint32(data))
	return b
}

// noDevice stands in for a machine whose audio device cannot open,
// counting open attempts and capturing the log.
func noDevice(t *testing.T) (opens *int, logs *[]string) {
	t.Helper()
	oldNew, oldLog := newContext, logf
	opens, logs = new(int), new([]string)
	newContext = func(int) (*audio.Context, error) {
		*opens++
		return nil, errors.New("oto: ALSA error at snd_pcm_open: \"default\": No such file or directory")
	}
	logf = func(format string, args ...any) { *logs = append(*logs, fmt.Sprintf(format, args...)) }
	ctxOnce, ctx = sync.Once{}, nil
	SetFS(fstest.MapFS{"hit.wav": {Data: tinyWAV()}, "theme.wav": {Data: tinyWAV()}})
	t.Cleanup(func() {
		newContext, logf = oldNew, oldLog
		ctxOnce, ctx = sync.Once{}, nil
		SetFS(nil)
	})
	return opens, logs
}

func TestNoAudioDevicePlaysSilently(t *testing.T) {
	opens, logs := noDevice(t)
	c := NewCache()
	for range 3 {
		c.PlaySound("hit.wav")
		if p := c.PlayMusic("theme.wav"); p != nil {
			t.Fatal("music without a device should return no player")
		}
	}
	if *opens != 1 {
		t.Fatalf("the device should be tried once, got %d", *opens)
	}
	if len(*logs) != 1 || !strings.Contains((*logs)[0], "no audio device") ||
		!strings.Contains((*logs)[0], "snd_pcm_open") {
		t.Fatalf("want one log line naming the cause, got %q", *logs)
	}
}

func TestNoAudioDeviceStillFailsOnMissingFiles(t *testing.T) {
	noDevice(t)
	c := NewCache() // shared: a failed load must leave it usable
	for _, play := range []func(string){c.PlaySound, func(p string) { c.PlayMusic(p) }} {
		func() {
			defer func() {
				if r := recover(); r == nil || !strings.Contains(fmt.Sprint(r), "cannot open audio file") {
					t.Fatalf("a missing file is a bug on every machine, got %v", r)
				}
			}()
			play("missing.wav")
		}()
	}
}
