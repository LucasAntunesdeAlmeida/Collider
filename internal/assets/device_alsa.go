//go:build ((linux && !android) || freebsd || netbsd) && !nintendosdk && !playstation5

package assets

// oto's ALSA fallback, probed the way oto opens it: libasound loaded at
// run time (purego, no cgo), the same device candidates in the same
// order, and the first that opens must take oto's format.

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"golang.org/x/sys/unix"
)

func init() { probeALSA = probeALSADevices }

const (
	sndPCMStreamPlayback = 0
	sndPCMNonblock       = 1
	sndPCMFormatFloatLE  = 14
	sndPCMAccessRWInter  = 3
	sndLatencyMicros     = 100000
)

var (
	alsaOnce sync.Once
	alsaErr  error

	sndStrerror          func(errnum int32) string
	sndPCMOpen           func(pcm *uintptr, name string, stream, mode int32) int32
	sndPCMSetParams      func(pcm uintptr, format, access int32, channels, rate uint32, resample int32, latency uint32) int32
	sndPCMClose          func(pcm uintptr) int32
	sndDeviceNameHint    func(card int32, iface string, hints *unsafe.Pointer) int32
	sndDeviceNameGetHint func(hint unsafe.Pointer, id string) unsafe.Pointer
	sndDeviceNameFree    func(hints unsafe.Pointer) int32
	libcFree             func(p unsafe.Pointer)
)

// loadALSA opens libasound the way oto does. A machine without it has
// no ALSA device, which is an answer, not a crash.
func loadALSA() error {
	alsaOnce.Do(func() {
		var h uintptr
		for _, name := range []string{"libasound.so.2", "libasound.so"} {
			if h, alsaErr = purego.Dlopen(name, purego.RTLD_LAZY|purego.RTLD_GLOBAL); alsaErr == nil {
				break
			}
		}
		if alsaErr != nil {
			alsaErr = fmt.Errorf("cannot load libasound: %v", alsaErr)
			return
		}
		purego.RegisterLibFunc(&sndStrerror, h, "snd_strerror")
		purego.RegisterLibFunc(&sndPCMOpen, h, "snd_pcm_open")
		purego.RegisterLibFunc(&sndPCMSetParams, h, "snd_pcm_set_params")
		purego.RegisterLibFunc(&sndPCMClose, h, "snd_pcm_close")
		purego.RegisterLibFunc(&sndDeviceNameHint, h, "snd_device_name_hint")
		purego.RegisterLibFunc(&sndDeviceNameGetHint, h, "snd_device_name_get_hint")
		purego.RegisterLibFunc(&sndDeviceNameFree, h, "snd_device_name_free_hint")
		// libc's free, for the hint strings: libasound links libc, so
		// its handle resolves it.
		purego.RegisterLibFunc(&libcFree, h, "free")
	})
	return alsaErr
}

// probeALSADevices tries oto's candidates until one opens and takes
// the format (a device that is merely busy counts: oto waits for it).
// ALSA and its plugins (JACK, OSS...) print every failed attempt to
// stderr; the engine logs one line instead, so stderr is muted while
// it looks.
func probeALSADevices(sampleRate int) error {
	if err := loadALSA(); err != nil {
		return err
	}
	defer quietStderr()()
	names := alsaCandidates(alsaHints())
	var first error
	for _, name := range names {
		var pcm uintptr
		if e := sndPCMOpen(&pcm, name, sndPCMStreamPlayback, sndPCMNonblock); e == -int32(unix.EBUSY) {
			return nil
		} else if e < 0 {
			if first == nil {
				first = fmt.Errorf("%q: %s", name, sndStrerror(e))
			}
			continue
		}
		e := sndPCMSetParams(pcm, sndPCMFormatFloatLE, sndPCMAccessRWInter, 2, uint32(sampleRate), 1, sndLatencyMicros)
		sndPCMClose(pcm)
		if e >= 0 {
			return nil
		}
		if first == nil {
			first = fmt.Errorf("%q cannot be configured: %s", name, sndStrerror(e))
		}
	}
	return fmt.Errorf("no playback device among %d (%v)", len(names), first)
}

// alsaHint is one PCM device ALSA lists: its name and direction
// ("Input", "Output", or "" for both).
type alsaHint struct{ name, ioid string }

// alsaHints lists ALSA's PCM devices, nil when it cannot.
func alsaHints() []alsaHint {
	var hints unsafe.Pointer
	if sndDeviceNameHint(-1, "pcm", &hints) != 0 {
		return nil
	}
	defer sndDeviceNameFree(hints)
	var out []alsaHint
	for p := hints; ; p = unsafe.Add(p, unsafe.Sizeof(uintptr(0))) {
		h := *(*unsafe.Pointer)(p)
		if h == nil {
			return out
		}
		out = append(out, alsaHint{name: hintString(h, "NAME"), ioid: hintString(h, "IOID")})
	}
}

// hintString is a hint's value, freeing the C string ALSA made.
func hintString(hint unsafe.Pointer, id string) string {
	p := sndDeviceNameGetHint(hint, id)
	if p == nil {
		return ""
	}
	defer libcFree(p)
	return unix.BytePtrToString((*byte)(p))
}

// alsaCandidates is oto's device order: default, plug:default, then
// every listed device that can play except null (and default again),
// each followed by its plug: variant, which converts the float samples
// to what the device takes.
func alsaCandidates(hints []alsaHint) []string {
	names := []string{"default", "plug:default"}
	for _, h := range hints {
		if h.ioid == "Input" {
			continue
		}
		switch h.name {
		case "", "null", "default":
			continue
		}
		names = append(names, h.name)
		if !strings.HasPrefix(h.name, "plug:") && !strings.HasPrefix(h.name, "plughw:") {
			names = append(names, "plug:"+h.name)
		}
	}
	return names
}

// quietStderr points file descriptor 2 at the null device until the
// returned function puts it back. Best effort: when it cannot, stderr
// stays as it was.
func quietStderr() (restore func()) {
	null, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		return func() {}
	}
	defer null.Close()
	saved, err := unix.Dup(2)
	if err != nil {
		return func() {}
	}
	if err := unix.Dup2(int(null.Fd()), 2); err != nil {
		unix.Close(saved)
		return func() {}
	}
	return func() {
		unix.Dup2(saved, 2)
		unix.Close(saved)
	}
}
