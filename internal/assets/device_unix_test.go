//go:build !android && !darwin && !js && !windows && !nintendosdk && !playstation5

package assets

import (
	"strings"
	"testing"
	"time"
)

// The probe must answer, never hang or crash, whether or not this
// machine has a sound card or a sound server (CI runners have neither).
func TestProbeDeviceAnswers(t *testing.T) {
	err := probeDevice(SampleRate)
	t.Logf("audio device: %v", err)
}

// No sound server is an answer, and a quick one: the probe falls back
// to ALSA instead of waiting.
func TestProbePulseWithoutServer(t *testing.T) {
	t.Setenv("PULSE_SERVER", "unix:"+t.TempDir()+"/no-server")
	start := time.Now()
	if err := probePulse(SampleRate); err == nil {
		t.Fatal("probePulse found a server behind a socket that does not exist")
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("probePulse took %v to give up", d)
	}
}

// Without a server and without ALSA the error names both, in oto's order.
func TestProbeDeviceNamesBoth(t *testing.T) {
	t.Setenv("PULSE_SERVER", "unix:"+t.TempDir()+"/no-server")
	saved := probeALSA
	t.Cleanup(func() { probeALSA = saved })
	probeALSA = func(int) error { return errNoALSA }
	err := probeDevice(SampleRate)
	if err == nil || !strings.HasPrefix(err.Error(), "PulseAudio: ") || !strings.HasSuffix(err.Error(), "; ALSA: no ALSA here") {
		t.Fatalf("probeDevice = %v, want PulseAudio then ALSA", err)
	}
	probeALSA = func(int) error { return nil }
	if err := probeDevice(SampleRate); err != nil {
		t.Fatalf("probeDevice = %v with a working ALSA device", err)
	}
}

type probeErr string

func (e probeErr) Error() string { return string(e) }

const errNoALSA = probeErr("no ALSA here")
