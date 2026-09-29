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

	// volumes holds each channel's volume, 0..1: the master one, and
	// the music and sound effect ones it multiplies (see Level).
	volumes [channels]float64
}

// Channel is a volume channel: Master scales everything, Music and
// Sounds scale their own kind under it.
type Channel int

const (
	Master Channel = iota
	Music
	Sounds
	channels
)

func NewCache() *Cache {
	return &Cache{
		images:  map[string]*ebiten.Image{},
		files:   map[string][]byte{},
		sounds:  map[string][]byte{},
		volumes: [channels]float64{1, 1, 1},
	}
}

// SetVolume sets a channel's volume, clamped to 0..1 (NaN = 0), for
// every player started from now on, and returns the stored value.
func (c *Cache) SetVolume(ch Channel, v float64) float64 {
	v = min(1, max(0, v))
	if v != v { // NaN
		v = 0
	}
	c.mu.Lock()
	c.volumes[ch] = v
	c.mu.Unlock()
	return v
}

// Volume returns a channel's own volume, as set.
func (c *Cache) Volume(ch Channel) float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.volumes[ch]
}

// Level returns the volume a player of this channel plays at: its own
// volume times the master one.
func (c *Cache) Level(ch Channel) float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.levelLocked(ch)
}

func (c *Cache) levelLocked(ch Channel) float64 {
	if ch == Master {
		return c.volumes[Master]
	}
	return c.volumes[Master] * c.volumes[ch]
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
