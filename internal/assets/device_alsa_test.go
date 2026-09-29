//go:build cgo && !android && !darwin && !js && !windows && !nintendosdk && !playstation5

package assets

import "testing"

// The probe must answer, never hang or crash, whether or not this
// machine has a sound card (CI runners have none).
func TestProbeDeviceAnswers(t *testing.T) {
	err := probeDevice(SampleRate)
	t.Logf("audio device: %v", err)
}
