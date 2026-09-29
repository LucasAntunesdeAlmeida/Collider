//go:build !cgo || android || darwin || js || windows || nintendosdk || playstation5

package assets

// probeDevice reports why audio cannot play here, or nil when it can.
// Outside ALSA the audio driver copes on its own: on Windows a machine
// with no output device gets a silent null device, and browsers,
// macOS and phones always have one.
func probeDevice(int) error { return nil }
