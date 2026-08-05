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

	"golang.org/x/image/font/gofont/gobold"
)

const sampleRate = 48000

func main() {
	// Pong
	writeWav("examples/pong/audios/bounce.wav", tone(440, 0.08, 0.5))

	// Jumper
	writeRect("examples/jumper/sprites/player.png", 32, 48, rgb(0x3E6FE5))
	writeCircle("examples/jumper/sprites/coin.png", 24, rgb(0xF2C938))
	writeSpikes("examples/jumper/sprites/spikes.png", 48, 24, rgb(0xE53E3E))
	writeRect("examples/jumper/sprites/replay.png", 140, 44, rgb(0x3EB658))
	writeRect("examples/jumper/sprites/retry.png", 140, 44, rgb(0xF28C38))
	writeWav("examples/jumper/audios/coin.wav", seq(tone(880, 0.06, 0.4), tone(1320, 0.1, 0.4)))
	writeWav("examples/jumper/audios/level1.wav", melody(0.22, 0.25,
		392, 440, 494, 587, 494, 440, 392, 330))

	// Zombie Night
	writeRect("examples/zombie-night/sprites/hero.png", 32, 32, rgb(0x3E6FE5))
	writeRect("examples/zombie-night/sprites/zombie.png", 32, 32, rgb(0x3EB658))
	writeRect("examples/zombie-night/sprites/again.png", 140, 44, rgb(0xF28C38))
	writeWav("examples/zombie-night/audios/shot.wav", noise(0.1, 0.4))
	writeWav("examples/zombie-night/audios/dread.wav", melody(0.5, 0.18,
		110, 104, 110, 98))

	// Breakout
	writeRect("examples/breakout/sprites/start.png", 160, 48, rgb(0x3EB658))
	writeWav("examples/breakout/audios/break.wav", tone(660, 0.07, 0.5))
	writeWav("examples/breakout/audios/menu.wav", melody(0.3, 0.2,
		262, 330, 392, 330))
	writeWav("examples/breakout/audios/action.wav", melody(0.16, 0.22,
		392, 392, 440, 494, 440, 392, 587, 494))

	// Caves
	writeWav("examples/caves/audios/gem.wav", seq(tone(1047, 0.05, 0.35), tone(1568, 0.08, 0.35)))

	// Runner: animation strips (horizontal rows of 32x48 frames).
	writeStrip("examples/runner/sprites/run.png", 4, 32, 48, drawRunFrame)
	writeStrip("examples/runner/sprites/jump.png", 2, 32, 48, drawJumpFrame)
	writeStrip("examples/runner/sprites/death.png", 4, 32, 48, drawDeathFrame)
	writeRect("examples/runner/sprites/crate.png", 30, 30, rgb(0x8B5A2B))
	writeWav("examples/runner/audios/jump.wav", seq(tone(330, 0.05, 0.35), tone(494, 0.08, 0.35)))
	writeWav("examples/runner/audios/hit.wav", noise(0.15, 0.45))
	writeWav("examples/runner/audios/run.wav", melody(0.14, 0.2,
		330, 392, 440, 392, 494, 440, 392, 330))

	// Dialog: portraits, theme, and a real bold TTF from the Go fonts.
	writeRect("examples/dialog/sprites/elder.png", 96, 96, rgb(0x9B59B6))
	writeRect("examples/dialog/sprites/hero.png", 96, 96, rgb(0x3E6FE5))
	writeFile("examples/dialog/fonts/gobold.ttf", gobold.TTF)
	writeWav("examples/dialog/audios/theme.wav", melody(0.5, 0.14,
		294, 330, 392, 330, 294, 262))

	// Dungeon
	writeWav("examples/dungeon/audios/dungeon.wav", melody(0.45, 0.14,
		147, 165, 131, 165))
	writeWav("examples/dungeon/audios/key.wav", seq(tone(784, 0.06, 0.35), tone(1047, 0.09, 0.35)))
	writeWav("examples/dungeon/audios/door.wav", seq(tone(196, 0.12, 0.4), noise(0.08, 0.2)))
	writeWav("examples/dungeon/audios/win.wav", melody(0.12, 0.3,
		523, 659, 784, 1047))

	// Catcher
	writeWav("examples/catcher/audios/catch.wav", seq(tone(880, 0.04, 0.35), tone(1175, 0.07, 0.35)))

	// Ship: hand-drawn pixel art (ASCII grids below), rendered at scale.
	writePixelStrip("examples/ship/sprites/ship.png", 3, shipPalette, shipFly)
	writePixelStrip("examples/ship/sprites/boom.png", 3, shipPalette, shipBoom)
	writePixelArt("examples/ship/sprites/comet.png", 3, shipPalette, cometArt)
	writePixelArt("examples/ship/sprites/icon.png", 2, shipPalette, shipFly[0])
	writeStars("examples/ship/sprites/background.png", 800, 600)
	writeBevel("examples/ship/sprites/button.png", 110, 30, 2)
	writeBevel("examples/ship/sprites/panel.png", 260, 130, 2)
	writeWav("examples/ship/audios/hit.wav", noise(0.18, 0.45))
	writeWav("examples/ship/audios/theme.wav", melody(0.2, 0.18,
		262, 330, 392, 523, 392, 330))

	// Memory
	writeRect("examples/memory/sprites/back.png", 100, 100, rgb(0x6B7280))
	faces := []uint32{
		0xE53E3E, 0x3EB658, 0x3E6FE5, 0xF2C938,
		0xF28C38, 0x9B59B6, 0x1ABC9C, 0xE91E8C,
	}
	for i, c := range faces {
		writeRect(fmt.Sprintf("examples/memory/sprites/face%d.png", i+1), 100, 100, rgb(c))
	}
	writeWav("examples/memory/audios/match.wav", seq(tone(523, 0.08, 0.4), tone(784, 0.12, 0.4)))
	writeWav("examples/memory/audios/calm.wav", melody(0.6, 0.15,
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

// shipPalette maps pixel-art runes to colors; space is transparent.
var shipPalette = map[rune]color.RGBA{
	'B': rgb(0x3E6FE5), // hull blue
	'C': rgb(0x9FD8FF), // cockpit glass
	'W': rgb(0xE8F0FF), // highlights
	'F': rgb(0xF28C38), // flame orange
	'Y': rgb(0xF2C938), // flame yellow
	'O': rgb(0xF28C38), // comet shell
	'S': rgb(0x666666), // smoke
}

// shipFly: two frames, the thruster flame alternates length.
var shipFly = [][]string{
	{
		"       BB       ",
		"      BBBB      ",
		"      BCCB      ",
		"     BBCCBB     ",
		"     BBCCBB     ",
		"    BBBBBBBB    ",
		"    BBBWWBBB    ",
		"   BBBBWWBBBB   ",
		"  BBBBBWWBBBBB  ",
		" WBBBBBBBBBBBBW ",
		" WBBBBBBBBBBBBW ",
		" WW BBBBBBBB WW ",
		"    BB    BB    ",
		"     F    F     ",
		"    FF    FF    ",
		"     Y    Y     ",
	},
	{
		"       BB       ",
		"      BBBB      ",
		"      BCCB      ",
		"     BBCCBB     ",
		"     BBCCBB     ",
		"    BBBBBBBB    ",
		"    BBBWWBBB    ",
		"   BBBBWWBBBB   ",
		"  BBBBBWWBBBBB  ",
		" WBBBBBBBBBBBBW ",
		" WBBBBBBBBBBBBW ",
		" WW BBBBBBBB WW ",
		"    BB    BB    ",
		"    FF    FF    ",
		"    YF    FY    ",
		"    YY    YY    ",
	},
}

// cometArt: a fireball with a sparking tail, falling downward.
var cometArt = []string{
	"     Y    Y     ",
	"      Y  Y      ",
	"      OYYO      ",
	"     OYYYYO     ",
	"    OYYWWYYO    ",
	"    OYWWWWYO    ",
	"   OYWWWWWWYO   ",
	"   OYWWWWWWYO   ",
	"   OYYWWWWYYO   ",
	"    OYYWWYYO    ",
	"    OOYYYYOO    ",
	"     OOYYOO     ",
	"      OOOO      ",
	"       OO       ",
	"                ",
	"                ",
}

// shipBoom: four frames, flash to ring to debris to smoke.
var shipBoom = [][]string{
	{
		"                ",
		"                ",
		"                ",
		"                ",
		"                ",
		"                ",
		"       YY       ",
		"      YWWY      ",
		"      YWWY      ",
		"       YY       ",
		"                ",
		"                ",
		"                ",
		"                ",
		"                ",
		"                ",
	},
	{
		"                ",
		"                ",
		"                ",
		"                ",
		"      YYYY      ",
		"     YOOOOY     ",
		"    YOWWWWOY    ",
		"    YOWWWWOY    ",
		"    YOWWWWOY    ",
		"     YOOOOY     ",
		"      YYYY      ",
		"                ",
		"                ",
		"                ",
		"                ",
		"                ",
	},
	{
		"                ",
		"     O    O     ",
		"    O OOOO O    ",
		"   O OYYYYO O   ",
		"     OYOOYO     ",
		"   O YO  OY O   ",
		"    OY    YO    ",
		"   O Y    Y O   ",
		"    OY    YO    ",
		"   O YO  OY O   ",
		"     OYOOYO     ",
		"   O OYYYYO O   ",
		"    O OOOO O    ",
		"     O    O     ",
		"                ",
		"                ",
	},
	{
		"                ",
		"                ",
		"   S       S    ",
		"                ",
		"      S  S      ",
		"                ",
		"  S    S     S  ",
		"                ",
		"     S    S     ",
		"                ",
		"  S     S    S  ",
		"                ",
		"       S        ",
		"                ",
		"                ",
		"                ",
	},
}

// writePixelArt renders one ASCII pixel grid at a scale factor.
func writePixelArt(path string, scale int, pal map[rune]color.RGBA, rows []string) {
	writePixelStrip(path, scale, pal, [][]string{rows})
}

// writePixelStrip renders ASCII pixel-art frames side by side into an
// animation strip.
func writePixelStrip(path string, scale int, pal map[rune]color.RGBA, frames [][]string) {
	h := len(frames[0])
	w := len(frames[0][0])
	img := image.NewRGBA(image.Rect(0, 0, w*scale*len(frames), h*scale))
	for f, rows := range frames {
		ox := f * w * scale
		for y, row := range rows {
			for x, r := range row {
				c, ok := pal[r]
				if !ok {
					continue
				}
				fillBox(img, ox+x*scale, y*scale, scale, scale, c)
			}
		}
	}
	writePng(path, img)
}

// writeStars paints a deep-space background with scattered stars.
func writeStars(path string, w, h int) {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	fillBox(img, 0, 0, w, h, rgb(0x0A0E1A))
	shades := []color.RGBA{rgb(0xFFFFFF), rgb(0x9FB6E8), rgb(0x5C6E96)}
	for range 170 {
		x, y := rand.IntN(w), rand.IntN(h)
		c := shades[rand.IntN(len(shades))]
		img.SetRGBA(x, y, c)
		if rand.IntN(4) == 0 {
			fillBox(img, x, y, 2, 2, c)
		}
	}
	writePng(path, img)
}

// writeBevel draws a pixel-style UI plate: mid fill, light top/left
// edge, dark bottom/right edge.
func writeBevel(path string, w, h, scale int) {
	sw, sh := w*scale, h*scale
	t := 2 * scale // edge thickness
	img := image.NewRGBA(image.Rect(0, 0, sw, sh))
	fillBox(img, 0, 0, sw, sh, rgb(0x2A3550))
	fillBox(img, 0, 0, sw, t, rgb(0x6E8BD8))
	fillBox(img, 0, 0, t, sh, rgb(0x6E8BD8))
	fillBox(img, 0, sh-t, sw, t, rgb(0x141B2E))
	fillBox(img, sw-t, 0, t, sh, rgb(0x141B2E))
	writePng(path, img)
}

// writeFile writes raw bytes (fonts), creating the folder if needed.
func writeFile(path string, data []byte) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		panic(err)
	}
	fmt.Println("  ", path)
}

// writeStrip draws an animation strip: frames horizontal slots of w x h,
// each painted by drawFrame(img, frame, originX).
func writeStrip(path string, frames, w, h int, drawFrame func(img *image.RGBA, f, ox int)) {
	img := image.NewRGBA(image.Rect(0, 0, w*frames, h))
	for f := range frames {
		drawFrame(img, f, f*w)
	}
	writePng(path, img)
}

// fillBox paints a solid rectangle onto an RGBA image.
func fillBox(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			if image.Pt(xx, yy).In(img.Rect) {
				img.SetRGBA(xx, yy, c)
			}
		}
	}
}

// drawRunFrame: a body with two legs whose offsets alternate per frame.
func drawRunFrame(img *image.RGBA, f, ox int) {
	body := rgb(0x3E6FE5)
	leg := rgb(0x1F3A80)
	fillBox(img, ox+8, 4, 16, 28, body)
	stride := []int{0, 5, 0, -5}[f%4]
	fillBox(img, ox+9+stride, 32, 5, 14, leg)
	fillBox(img, ox+19-stride, 32, 5, 14, leg)
}

// drawJumpFrame: legs tucked, slight bob between the two frames.
func drawJumpFrame(img *image.RGBA, f, ox int) {
	body := rgb(0x3E6FE5)
	leg := rgb(0x1F3A80)
	bob := f * 3
	fillBox(img, ox+8, 6+bob, 16, 28, body)
	fillBox(img, ox+9, 34+bob, 6, 8, leg)
	fillBox(img, ox+18, 34+bob, 6, 8, leg)
}

// drawDeathFrame: the body collapses and darkens frame by frame.
func drawDeathFrame(img *image.RGBA, f, ox int) {
	fade := uint8(0xE5 - f*40)
	c := color.RGBA{R: fade, G: 0x3E, B: 0x3E, A: 0xFF}
	drop := f * 8
	height := 28 - f*6
	fillBox(img, ox+4, 16+drop, 24, height, c)
}

// writeWav writes 16-bit mono PCM at 48kHz, creating the folder if needed.
func writeWav(path string, samples []int16) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		panic(err)
	}
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
