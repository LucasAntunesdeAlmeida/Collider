// Package record captures gameplay frames and writes an animated GIF.
// It exists so any Collider game can produce a shareable clip with zero
// tooling: set COLLIDER_RECORD=demo.gif, play, close the window.
package record

import (
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/gif"
	"os"
	"time"
)

const (
	scale      = 2                     // capture at half resolution
	frameEvery = 80 * time.Millisecond // ~12.5 fps
	gifDelay   = 8                     // in 1/100s, matches frameEvery
	maxFrames  = 900                   // ~72s, bounds memory
)

// Recorder accumulates downscaled palettized frames and encodes them on
// Save. All methods run on the game loop; nothing is concurrent.
type Recorder struct {
	path   string
	w, h   int
	last   time.Time
	frames []*image.Paletted

	// memo caches RGB -> palette index; games use few distinct colors,
	// so this turns palettization from a per-pixel search into a lookup.
	memo map[uint32]uint8
}

func New(path string, w, h int) *Recorder {
	return &Recorder{path: path, w: w, h: h, memo: map[uint32]uint8{}}
}

// ShouldCapture throttles capture to the GIF frame rate regardless of
// the display's refresh rate.
func (r *Recorder) ShouldCapture() bool {
	if len(r.frames) >= maxFrames {
		return false
	}
	if time.Since(r.last) < frameEvery {
		return false
	}
	r.last = time.Now()
	return true
}

// Add consumes one full-resolution RGBA frame (4 bytes per pixel).
func (r *Recorder) Add(pix []byte) {
	ow, oh := r.w/scale, r.h/scale
	img := image.NewPaletted(image.Rect(0, 0, ow, oh), palette.Plan9)
	for y := range oh {
		for x := range ow {
			off := (y*scale*r.w + x*scale) * 4
			key := uint32(pix[off])<<16 | uint32(pix[off+1])<<8 | uint32(pix[off+2])
			idx, ok := r.memo[key]
			if !ok {
				idx = uint8(color.Palette(palette.Plan9).Index(
					color.RGBA{R: pix[off], G: pix[off+1], B: pix[off+2], A: 0xFF}))
				r.memo[key] = idx
			}
			img.Pix[y*img.Stride+x] = idx
		}
	}
	r.frames = append(r.frames, img)
}

// Save encodes the captured frames as an animated GIF.
func (r *Recorder) Save() {
	if len(r.frames) == 0 {
		return
	}
	g := &gif.GIF{}
	for _, f := range r.frames {
		g.Image = append(g.Image, f)
		g.Delay = append(g.Delay, gifDelay)
	}
	f, err := os.Create(r.path)
	if err != nil {
		fmt.Println("collider: cannot save recording:", err)
		return
	}
	defer f.Close()
	if err := gif.EncodeAll(f, g); err != nil {
		fmt.Println("collider: cannot encode recording:", err)
		return
	}
	fmt.Printf("collider: recording saved to %s (%d frames)\n", r.path, len(r.frames))
}
