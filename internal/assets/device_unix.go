//go:build !android && !darwin && !js && !windows && !nintendosdk && !playstation5

package assets

// This is where Ebitengine's audio driver (oto) plays through
// PulseAudio (PipeWire speaks it too), falling back to ALSA. It opens
// the device in the background and a failure (no sound card, no sound
// server, as in a bare container) would end the game loop, so the
// engine tries what oto will try first, in the same order: the
// PulseAudio server, then the ALSA devices. No cgo: PulseAudio is a
// socket protocol spoken in Go, ALSA a library loaded at run time.

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jfreymuth/pulse"
)

// probeALSA reports why no ALSA device can play, or nil when one can.
// device_alsa.go sets it where oto falls back to ALSA; elsewhere (the
// BSDs without it) PulseAudio is the only way.
var probeALSA func(sampleRate int) error

// probeDevice reports why audio cannot play here, or nil when it can.
func probeDevice(sampleRate int) error {
	perr := probePulse(sampleRate)
	if perr == nil {
		return nil
	}
	if probeALSA == nil {
		return fmt.Errorf("PulseAudio: %v", perr)
	}
	aerr := probeALSA(sampleRate)
	if aerr == nil {
		return nil
	}
	return fmt.Errorf("PulseAudio: %v; ALSA: %v", perr, aerr)
}

// probePulse connects to the sound server and creates the stream oto
// will (stereo float at sampleRate), corked, so nothing plays, then
// closes both. A server that does not answer times out after the
// library's default second.
func probePulse(sampleRate int) error {
	name := "Oto" // the client name oto uses, so the server sees one app
	if exe, err := os.Executable(); err == nil {
		name = filepath.Base(exe)
	}
	c, err := pulse.NewClient(pulse.ClientApplicationName(name))
	if err != nil {
		return err
	}
	defer c.Close()
	silence := pulse.Float32Reader(func(buf []float32) (int, error) {
		clear(buf)
		return len(buf), nil
	})
	s, err := c.NewPlayback(silence, pulse.PlaybackStereo, pulse.PlaybackSampleRate(sampleRate))
	if err != nil {
		return err
	}
	s.Close()
	return nil
}
