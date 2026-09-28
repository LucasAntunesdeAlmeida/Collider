package main

import (
	"image"
	"image/color"
	"math"
	"math/rand/v2"
)

// pal is the shared pixel-art palette. Every sprite grid in this tool
// uses these runes; space is transparent.
var pal = map[rune]color.RGBA{
	// blues
	'b': rgb(0x1F3A80), 'B': rgb(0x3E6FE5), 'c': rgb(0x6E8BD8), 'C': rgb(0x9FD8FF),
	// greens
	'g': rgb(0x2C7A46), 'G': rgb(0x3EB658), 'h': rgb(0x7FE39A),
	// reds
	'r': rgb(0x8E2020), 'R': rgb(0xE53E3E), 'p': rgb(0xFF8A8A),
	// yellows and oranges
	'y': rgb(0xC79A18), 'Y': rgb(0xF2C938), 'o': rgb(0xF28C38), 'O': rgb(0xC26A18),
	// purples and pinks
	'm': rgb(0x6B3FA0), 'M': rgb(0x9B59B6), 'k': rgb(0xE91E8C),
	// neutrals
	'W': rgb(0xF2F2F2), 'w': rgb(0xC8CEDA), 's': rgb(0x8A93A6), 'S': rgb(0x4A5266),
	'd': rgb(0x2A3550), 'D': rgb(0x141B2E), 'n': rgb(0x0A0E1A),
	// browns
	't': rgb(0x6B4423), 'T': rgb(0x8B5A2B), 'u': rgb(0xB07C4A),
	// skin
	'f': rgb(0xE8B080), 'F': rgb(0xFFD3A8),
}

func rgb(v uint32) color.RGBA {
	return color.RGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 0xFF}
}

// art renders one ASCII grid to a PNG at the given scale.
func art(path string, scale int, rows []string) {
	strip(path, scale, [][]string{rows})
}

// strip renders ASCII grids side by side as an animation strip.
func strip(path string, scale int, frames [][]string) {
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

// backdrop paints a vertical gradient between two colors, the base for
// most game backgrounds.
func backdrop(path string, w, h int, top, bottom color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		f := float64(y) / float64(h-1)
		c := color.RGBA{
			R: uint8(float64(top.R)*(1-f) + float64(bottom.R)*f),
			G: uint8(float64(top.G)*(1-f) + float64(bottom.G)*f),
			B: uint8(float64(top.B)*(1-f) + float64(bottom.B)*f),
			A: 0xFF,
		}
		fillBox(img, 0, y, w, 1, c)
	}
	if path != "" {
		writePng(path, img)
	}
	return img
}

// starfield scatters stars over a backdrop.
func starfield(path string, w, h int, count int, top, bottom color.RGBA) {
	img := backdrop("", w, h, top, bottom)
	shades := []color.RGBA{rgb(0xFFFFFF), rgb(0x9FB6E8), rgb(0x5C6E96)}
	for range count {
		x, y := rand.IntN(w), rand.IntN(h)
		c := shades[rand.IntN(len(shades))]
		img.SetRGBA(x, y, c)
		if rand.IntN(5) == 0 {
			fillBox(img, x, y, 2, 2, c)
		}
	}
	writePng(path, img)
}

// tile draws a repeating textured block: the wall, floor and platform
// material of the tile-based games. Flecks scatter darker specks.
func tile(path string, w, h int, base, light, dark color.RGBA, flecks int) {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	fillBox(img, 0, 0, w, h, base)
	fillBox(img, 0, 0, w, 2, light)
	fillBox(img, 0, 0, 2, h, light)
	fillBox(img, 0, h-2, w, 2, dark)
	fillBox(img, w-2, 0, 2, h, dark)
	for range flecks {
		x, y := 3+rand.IntN(max(w-6, 1)), 3+rand.IntN(max(h-6, 1))
		fillBox(img, x, y, 2, 2, dark)
	}
	writePng(path, img)
}

// hills paints rolling ground silhouettes over a backdrop: the outdoor
// background for the platformers.
func hills(path string, w, h int, sky, ground color.RGBA) {
	img := backdrop("", w, h, sky, shade(sky, 1.25))
	far := shade(ground, 0.55)
	near := shade(ground, 0.8)
	for x := range w {
		fx := float64(x)
		y1 := float64(h)*0.62 + math.Sin(fx/110)*38 + math.Sin(fx/47)*12
		y2 := float64(h)*0.78 + math.Sin(fx/80+2)*26
		fillBox(img, x, int(y1), 1, h-int(y1), far)
		fillBox(img, x, int(y2), 1, h-int(y2), near)
	}
	writePng(path, img)
}

// shade scales a color's channels.
func shade(c color.RGBA, f float64) color.RGBA {
	clamp := func(v float64) uint8 {
		if v < 0 {
			return 0
		}
		if v > 255 {
			return 255
		}
		return uint8(v)
	}
	return color.RGBA{R: clamp(float64(c.R) * f), G: clamp(float64(c.G) * f), B: clamp(float64(c.B) * f), A: 0xFF}
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

// disc draws a pixel-art circle of cells x cells blocks at scale: a
// solid disc when hole is 0, else a ring whose inner radius is hole
// cells. The outer rim uses edge, the rest body. The horde example's
// virtual joystick is a ring and a knob.
func disc(path string, cells, scale int, hole float64, body, edge color.RGBA) {
	img := image.NewRGBA(image.Rect(0, 0, cells*scale, cells*scale))
	r := float64(cells) / 2
	for y := range cells {
		for x := range cells {
			d := math.Hypot(float64(x)+0.5-r, float64(y)+0.5-r)
			switch {
			case d > r || d < hole:
				continue
			case d > r-1.2 || (hole > 0 && d < hole+1):
				fillBox(img, x*scale, y*scale, scale, scale, edge)
			default:
				fillBox(img, x*scale, y*scale, scale, scale, body)
			}
		}
	}
	writePng(path, img)
}
