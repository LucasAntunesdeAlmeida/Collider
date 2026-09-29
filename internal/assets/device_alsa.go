//go:build cgo && !android && !darwin && !js && !windows && !nintendosdk && !playstation5

package assets

// This is where Ebitengine's audio driver (oto) plays through ALSA. It
// opens the device in the background and a failure (no sound card, no
// PulseAudio or PipeWire, as in a bare container) would end the game
// loop, so the engine tries the same devices oto does first, the same
// way: the first one that opens must take oto's format.

// #cgo pkg-config: alsa
//
// #include <alsa/asoundlib.h>
// #include <errno.h>
// #include <stdlib.h>
// #include <string.h>
//
// // ALSA prints every failed attempt to stderr; the engine logs one line.
// static void colliderQuiet(const char *file, int line, const char *fn, int err, const char *fmt, ...) {}
//
// // Returns the open error (< 0) when name cannot be opened, else 0 with
// // *format set to the result of configuring it the way oto will.
// static int colliderTry(const char *name, unsigned int rate, int *format) {
// 	snd_pcm_t *pcm;
// 	int err = snd_pcm_open(&pcm, name, SND_PCM_STREAM_PLAYBACK, SND_PCM_NONBLOCK);
// 	if (err == -EBUSY) {
// 		*format = 0; // there, just in use: oto waits for it
// 		return 0;
// 	}
// 	if (err < 0) {
// 		return err;
// 	}
// 	*format = snd_pcm_set_params(pcm, SND_PCM_FORMAT_FLOAT_LE, SND_PCM_ACCESS_RW_INTERLEAVED, 2, rate, 1, 100000);
// 	snd_pcm_close(pcm);
// 	return 0;
// }
//
// // oto's candidates, in its order: default, plug:default, then every
// // output device ALSA lists except null.
// static int colliderProbe(unsigned int rate, int *format) {
// 	snd_lib_error_set_handler(colliderQuiet);
// 	int err = colliderTry("default", rate, format);
// 	if (err < 0) {
// 		err = colliderTry("plug:default", rate, format);
// 	}
// 	void **hints;
// 	if (err < 0 && snd_device_name_hint(-1, "pcm", &hints) == 0) {
// 		for (void **it = hints; *it != NULL && err < 0; it++) {
// 			char *io = snd_device_name_get_hint(*it, "IOID");
// 			char *name = snd_device_name_get_hint(*it, "NAME");
// 			if (name != NULL && (io == NULL || strcmp(io, "Input") != 0) &&
// 				strcmp(name, "null") != 0 && strcmp(name, "default") != 0) {
// 				err = colliderTry(name, rate, format);
// 			}
// 			free(io);
// 			free(name);
// 		}
// 		snd_device_name_free_hint(hints);
// 	}
// 	snd_lib_error_set_handler(NULL);
// 	return err;
// }
import "C"

import "fmt"

// probeDevice reports why audio cannot play here, or nil when it can.
func probeDevice(sampleRate int) error {
	var format C.int
	if err := C.colliderProbe(C.uint(sampleRate), &format); err < 0 {
		return fmt.Errorf("ALSA cannot open a playback device: %s", C.GoString(C.snd_strerror(err)))
	}
	if format < 0 {
		return fmt.Errorf("ALSA playback device cannot be configured: %s", C.GoString(C.snd_strerror(format)))
	}
	return nil
}
