// Command genassets writes placeholder sprites (PNG) and sounds (WAV)
// into the example folders, so every example runs out of the box without
// shipping binary art anyone has to draw. Run from the repo root:
//
//	go run ./tools/genassets
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
)

const sampleRate = 48000

func main() {
	// Pong
	writeWav("examples/pong/bounce.wav", tone(440, 0.08, 0.5))

	// Jumper
	writeRect("examples/jumper/player.png", 32, 48, rgb(0x3E6FE5))
	writeCircle("examples/jumper/coin.png", 24, rgb(0xF2C938))
	writeSpikes("examples/jumper/spikes.png", 48, 24, rgb(0xE53E3E))
	writeRect("examples/jumper/replay.png", 140, 44, rgb(0x3EB658))
	writeRect("examples/jumper/retry.png", 140, 44, rgb(0xF28C38))
	writeWav("examples/jumper/coin.wav", seq(tone(880, 0.06, 0.4), tone(1320, 0.1, 0.4)))
	writeWav("examples/jumper/level1.wav", melody(0.22, 0.25,
		392, 440, 494, 587, 494, 440, 392, 330))

	// Zombie Night
	writeRect("examples/zombie-night/hero.png", 32, 32, rgb(0x3E6FE5))
	writeRect("examples/zombie-night/zombie.png", 32, 32, rgb(0x3EB658))
	writeRect("examples/zombie-night/again.png", 140, 44, rgb(0xF28C38))
	writeWav("examples/zombie-night/shot.wav", noise(0.1, 0.4))
	writeWav("examples/zombie-night/dread.wav", melody(0.5, 0.18,
		110, 104, 110, 98))

	// Breakout
	writeRect("examples/breakout/start.png", 160, 48, rgb(0x3EB658))
	writeWav("examples/breakout/break.wav", tone(660, 0.07, 0.5))
	writeWav("examples/breakout/menu.wav", melody(0.3, 0.2,
		262, 330, 392, 330))
	writeWav("examples/breakout/action.wav", melody(0.16, 0.22,
		392, 392, 440, 494, 440, 392, 587, 494))

	// Memory
	writeRect("examples/memory/back.png", 100, 100, rgb(0x6B7280))
	faces := []uint32{
		0xE53E3E, 0x3EB658, 0x3E6FE5, 0xF2C938,
		0xF28C38, 0x9B59B6, 0x1ABC9C, 0xE91E8C,
	}
	for i, c := range faces {
		writeRect(fmt.Sprintf("examples/memory/face%d.png", i+1), 100, 100, rgb(c))
	}
	writeWav("examples/memory/match.wav", seq(tone(523, 0.08, 0.4), tone(784, 0.12, 0.4)))
	writeWav("examples/memory/calm.wav", melody(0.6, 0.15,
		262, 294, 330, 294))

	fmt.Println("assets generated")
}

func rgb(v uint32) color.RGBA {
	return color.RGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 0xFF}
}

// writePng encodes and writes an image, creating the folder if needed.
func writePng(path string, img image.Image) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		panic(err)
	}
	fmt.Println("  ", path)
}

// writeRect draws a filled rectangle with a darker 2px border.
func writeRect(path string, w, h int, c color.RGBA) {
	border := color.RGBA{R: c.R / 2, G: c.G / 2, B: c.B / 2, A: 0xFF}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			if x < 2 || y < 2 || x >= w-2 || y >= h-2 {
				img.SetRGBA(x, y, border)
			} else {
				img.SetRGBA(x, y, c)
			}
		}
	}
	writePng(path, img)
}

// writeCircle draws a filled circle on a transparent background.
func writeCircle(path string, d int, c color.RGBA) {
	img := image.NewRGBA(image.Rect(0, 0, d, d))
	r := float64(d) / 2
	for y := range d {
		for x := range d {
			dx := float64(x) + 0.5 - r
			dy := float64(y) + 0.5 - r
			if math.Hypot(dx, dy) <= r {
				img.SetRGBA(x, y, c)
			}
		}
	}
	writePng(path, img)
}

// writeSpikes draws a row of triangles on a transparent background.
func writeSpikes(path string, w, h int, c color.RGBA) {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	period := float64(w) / 3
	for y := range h {
		for x := range w {
			phase := math.Mod(float64(x), period) / period
			peak := 1 - math.Abs(2*phase-1) // 0 at edges, 1 at triangle tip
			if float64(h-y) <= peak*float64(h) {
				img.SetRGBA(x, y, c)
			}
		}
	}
	writePng(path, img)
}

// writeWav writes 16-bit mono PCM at 48kHz.
func writeWav(path string, samples []int16) {
	var buf bytes.Buffer
	dataLen := len(samples) * 2
	buf.WriteString("RIFF")
	binary.Write(&buf, binary.LittleEndian, uint32(36+dataLen))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	binary.Write(&buf, binary.LittleEndian, uint32(16))
	binary.Write(&buf, binary.LittleEndian, uint16(1)) // PCM
	binary.Write(&buf, binary.LittleEndian, uint16(1)) // mono
	binary.Write(&buf, binary.LittleEndian, uint32(sampleRate))
	binary.Write(&buf, binary.LittleEndian, uint32(sampleRate*2))
	binary.Write(&buf, binary.LittleEndian, uint16(2))
	binary.Write(&buf, binary.LittleEndian, uint16(16))
	buf.WriteString("data")
	binary.Write(&buf, binary.LittleEndian, uint32(dataLen))
	binary.Write(&buf, binary.LittleEndian, samples)
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		panic(err)
	}
	fmt.Println("  ", path)
}

// tone is a sine burst with a linear decay envelope.
func tone(freq, sec, vol float64) []int16 {
	n := int(sec * sampleRate)
	out := make([]int16, n)
	for i := range n {
		t := float64(i) / sampleRate
		env := 1 - float64(i)/float64(n)
		out[i] = int16(vol * env * 32000 * math.Sin(2*math.Pi*freq*t))
	}
	return out
}

// noise is a white-noise burst with decay, for shots and hits.
func noise(sec, vol float64) []int16 {
	n := int(sec * sampleRate)
	out := make([]int16, n)
	for i := range n {
		env := 1 - float64(i)/float64(n)
		out[i] = int16(vol * env * 32000 * (rand.Float64()*2 - 1))
	}
	return out
}

// melody plays notes in sequence; loops cleanly as background music.
func melody(noteSec, vol float64, freqs ...float64) []int16 {
	var out []int16
	for _, f := range freqs {
		out = append(out, tone(f, noteSec, vol)...)
	}
	return out
}

// seq concatenates sample runs.
func seq(parts ...[]int16) []int16 {
	var out []int16
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}
