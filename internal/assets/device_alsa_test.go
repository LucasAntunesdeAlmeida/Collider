//go:build ((linux && !android) || freebsd || netbsd) && !nintendosdk && !playstation5

package assets

import (
	"os"
	"reflect"
	"testing"

	"golang.org/x/sys/unix"
)

// The ALSA probe must answer, never hang or crash, with or without
// libasound and a sound card.
func TestProbeALSAAnswers(t *testing.T) {
	err := probeALSADevices(SampleRate)
	t.Logf("ALSA device: %v", err)
}

// oto's order: default, plug:default, then each output device and its
// plug: variant; inputs, null and default itself are skipped.
func TestALSACandidates(t *testing.T) {
	got := alsaCandidates([]alsaHint{
		{name: "null"},
		{name: "default"},
		{name: "sysdefault:CARD=PCH"},
		{name: "mic", ioid: "Input"},
		{name: "hdmi:CARD=HDMI,DEV=0", ioid: "Output"},
		{name: "plughw:CARD=PCH,DEV=0"},
		{name: "plug:dmix"},
		{name: ""},
	})
	want := []string{
		"default", "plug:default",
		"sysdefault:CARD=PCH", "plug:sysdefault:CARD=PCH",
		"hdmi:CARD=HDMI,DEV=0", "plug:hdmi:CARD=HDMI,DEV=0",
		"plughw:CARD=PCH,DEV=0",
		"plug:dmix",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("alsaCandidates =\n%q\nwant\n%q", got, want)
	}
	if got := alsaCandidates(nil); !reflect.DeepEqual(got, []string{"default", "plug:default"}) {
		t.Fatalf("alsaCandidates(nil) = %q", got)
	}
}

// Muting stderr points fd 2 at the null device, and restoring puts the
// original file back.
func TestQuietStderr(t *testing.T) {
	var before, during, after, null unix.Stat_t
	if err := unix.Fstat(2, &before); err != nil {
		t.Skip("no stderr:", err)
	}
	if err := unix.Stat(os.DevNull, &null); err != nil {
		t.Skip("no null device:", err)
	}
	restore := quietStderr()
	err := unix.Fstat(2, &during)
	restore()
	if err != nil {
		t.Fatal(err)
	}
	if during.Dev != null.Dev || during.Ino != null.Ino {
		t.Fatal("stderr is not the null device while muted")
	}
	if err := unix.Fstat(2, &after); err != nil {
		t.Fatal(err)
	}
	if after.Dev != before.Dev || after.Ino != before.Ino {
		t.Fatal("stderr was not restored")
	}
}
