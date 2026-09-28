// Package assets loads and caches game content: images, sound effects
// and music.
package assets

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// Cache loads and memoizes content by file path. Each Game owns one,
// so the same file is decoded and uploaded to the GPU only once.
type Cache struct {
	mu     sync.Mutex
	images map[string]*ebiten.Image
	files  map[string][]byte
	sounds map[string][]byte

	// volume is the master volume, 0..1, applied to every player the
	// cache starts.
	volume float64
}

func NewCache() *Cache {
	return &Cache{
		images: map[string]*ebiten.Image{},
		files:  map[string][]byte{},
		sounds: map[string][]byte{},
		volume: 1,
	}
}

// SetVolume sets the master volume, clamped to 0..1, for every sound
// and music player started from now on, and returns the stored value.
func (c *Cache) SetVolume(v float64) float64 {
	v = min(1, max(0, v))
	if v != v { // NaN
		v = 0
	}
	c.mu.Lock()
	c.volume = v
	c.mu.Unlock()
	return v
}

// Volume returns the master volume.
func (c *Cache) Volume() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.volume
}

// Image returns the cached image for a path, loading it on first use.
// It panics with a clear message on a missing or broken file: a game
// with a missing sprite is a programming bug, not a runtime condition.
func (c *Cache) Image(path string) *ebiten.Image {
	c.mu.Lock()
	defer c.mu.Unlock()
	if img, ok := c.images[path]; ok {
		return img
	}
	b, err := ReadFile(path)
	if err != nil {
		panic("collider: cannot open sprite " + path + ": " + err.Error())
	}
	src, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		panic("collider: cannot decode sprite " + path + ": " + err.Error())
	}
	img := ebiten.NewImageFromImage(src)
	c.images[path] = img
	return img
}

// fileLocked returns the raw bytes of a file, cached by path. The
// caller must hold c.mu.
func (c *Cache) fileLocked(path string) []byte {
	if b, ok := c.files[path]; ok {
		return b
	}
	b, err := ReadFile(path)
	if err != nil {
		panic("collider: cannot open audio file " + path + ": " + err.Error())
	}
	c.files[path] = b
	return b
}
