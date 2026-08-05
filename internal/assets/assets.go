// Package assets loads and caches game content: images, sound effects
// and music.
package assets

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
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
}

func NewCache() *Cache {
	return &Cache{
		images: map[string]*ebiten.Image{},
		files:  map[string][]byte{},
		sounds: map[string][]byte{},
	}
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
	f, err := os.Open(path)
	if err != nil {
		panic("collider: cannot open sprite " + path + ": " + err.Error())
	}
	defer f.Close()
	src, _, err := image.Decode(f)
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
	b, err := os.ReadFile(path)
	if err != nil {
		panic("collider: cannot open audio file " + path + ": " + err.Error())
	}
	c.files[path] = b
	return b
}
