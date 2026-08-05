// Command cover generates the repository's social preview image
// (.github/social-preview.png, 1280x640) from real project assets: the
// pixel font, the engine palette and gameplay frames pulled straight
// from the example demo GIFs. Run from the repo root:
//
//	go run ./tools/cover
package main

import (
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"log"
	"math/rand/v2"
	"os"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	coverW = 1280
	coverH = 640
)

func rgb(v uint32) color.RGBA {
	return color.RGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 0xFF}
}

var (
	night    = rgb(0x0A0E1A)
	dusk     = rgb(0x1A2340)
	yellow   = rgb(0xF2C938)
	white    = rgb(0xF2F2F2)
	grey     = rgb(0xC8CEDA)
	cardEdge = rgb(0x4A5266)
)

func main() {
	img := image.NewRGBA(image.Rect(0, 0, coverW, coverH))

	background(img)
	sprites(img)
	titleBlock(img)
	gameStrip(img)

	if err := os.MkdirAll(".github", 0o755); err != nil {
		log.Fatal(err)
	}
	out, err := os.Create(".github/social-preview.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := png.Encode(out, img); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote .github/social-preview.png")
}

// background paints the night gradient and a deterministic starfield,
// the same look as the ship example's sky.
func background(img *image.RGBA) {
	for y := range coverH {
		f := float64(y) / float64(coverH-1)
		c := color.RGBA{
			R: uint8(float64(night.R)*(1-f) + float64(dusk.R)*f),
			G: uint8(float64(night.G)*(1-f) + float64(dusk.G)*f),
			B: uint8(float64(night.B)*(1-f) + float64(dusk.B)*f),
			A: 0xFF,
		}
		fill(img, 0, y, coverW, 1, c)
	}
	rnd := rand.New(rand.NewPCG(7, 7))
	shades := []color.RGBA{rgb(0xFFFFFF), rgb(0x9FB6E8), rgb(0x5C6E96)}
	for range 340 {
		x, y := rnd.IntN(coverW), rnd.IntN(coverH)
		c := shades[rnd.IntN(len(shades))]
		img.SetRGBA(x, y, c)
		if rnd.IntN(5) == 0 {
			fill(img, x, y, 2, 2, c)
		}
	}
}

// sprites places the ship and a comet from the ship example as pixel
// decorations around the title.
func sprites(img *image.RGBA) {
	ship := loadPNG("examples/ship/sprites/ship.png")
	if ship != nil {
		b := ship.Bounds()
		frame := image.Rect(b.Min.X, b.Min.Y, b.Min.X+b.Dx()/2, b.Max.Y)
		pasteScaled(img, ship, frame, 1080, 90, 3)
	}
	comet := loadPNG("examples/ship/sprites/comet.png")
	if comet != nil {
		pasteScaled(img, comet, comet.Bounds(), 48, 140, 2)
		pasteScaled(img, comet, comet.Bounds(), 1150, 300, 2)
	}
}

// titleBlock draws the name and tagline in the project's pixel font.
func titleBlock(img *image.RGBA) {
	title := face(112)
	tag := face(30)

	center(img, title, "COLLIDER", 178, yellow, true)
	center(img, tag, "2D GAMES IN GO", 258, white, false)
	center(img, tag, "PLAYED BY HUMANS AND AI AGENTS", 306, grey, false)
}

// gameStrip lays a row of five gameplay frames, pulled live from the
// example demo GIFs, along the bottom of the cover.
func gameStrip(img *image.RGBA) {
	games := []struct {
		name string
		at   float64 // how far into the recording to sample, 0..1
	}{
		{"zombie-night", 0.8}, {"breakout", 0.35}, {"runner", 0.5},
		{"dungeon", 0.5}, {"agent", 0.55},
	}
	const (
		cardW = 232
		cardH = 174
		gap   = 16
		y     = 408
	)
	total := len(games)*cardW + (len(games)-1)*gap
	x := (coverW - total) / 2
	for _, g := range games {
		frame := gifFrame("examples/"+g.name+"/demo.gif", g.at)
		if frame == nil {
			x += cardW + gap
			continue
		}
		fill(img, x-3, y-3, cardW+6, cardH+6, cardEdge)
		fill(img, x-1, y-1, cardW+2, cardH+2, night)
		dst := image.Rect(x, y, x+cardW, y+cardH)
		xdraw.CatmullRom.Scale(img, dst, frame, frame.Bounds(), xdraw.Over, nil)
		x += cardW + gap
	}
}

// gifFrame decodes a demo GIF and composites frames in order, honoring
// their offsets, returning the state the given fraction into the
// recording so the frame shows real mid-game action.
func gifFrame(path string, at float64) *image.RGBA {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("skip %s: %v", path, err)
		return nil
	}
	defer f.Close()
	g, err := gif.DecodeAll(f)
	if err != nil || len(g.Image) == 0 {
		log.Printf("skip %s: %v", path, err)
		return nil
	}
	target := int(float64(len(g.Image)-1) * at)
	out := image.NewRGBA(image.Rect(0, 0, g.Config.Width, g.Config.Height))
	for i := 0; i <= target; i++ {
		fr := g.Image[i]
		xdraw.Draw(out, fr.Bounds(), fr, fr.Bounds().Min, xdraw.Over)
	}
	return out
}

func face(size float64) font.Face {
	data, err := os.ReadFile("tools/genassets/fontsrc/pixel.ttf")
	if err != nil {
		log.Fatal(err)
	}
	ft, err := opentype.Parse(data)
	if err != nil {
		log.Fatal(err)
	}
	fc, err := opentype.NewFace(ft, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		log.Fatal(err)
	}
	return fc
}

// center draws a line of text horizontally centered at baseline y, with
// an optional hard pixel shadow.
func center(img *image.RGBA, fc font.Face, s string, y int, c color.RGBA, shadow bool) {
	d := &font.Drawer{Dst: img, Face: fc}
	w := d.MeasureString(s).Ceil()
	x := (coverW - w) / 2
	if shadow {
		d.Src = image.NewUniform(rgb(0x000000))
		d.Dot = fixed.P(x+6, y+6)
		d.DrawString(s)
	}
	d.Src = image.NewUniform(c)
	d.Dot = fixed.P(x, y)
	d.DrawString(s)
}

func loadPNG(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("skip %s: %v", path, err)
		return nil
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		log.Printf("skip %s: %v", path, err)
		return nil
	}
	return img
}

// pasteScaled draws src's sub-rectangle onto dst at (x, y), scaled by an
// integer factor with hard pixel edges.
func pasteScaled(dst *image.RGBA, src image.Image, sub image.Rectangle, x, y, scale int) {
	r := image.Rect(x, y, x+sub.Dx()*scale, y+sub.Dy()*scale)
	xdraw.NearestNeighbor.Scale(dst, r, src, sub, xdraw.Over, nil)
}

func fill(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			if image.Pt(xx, yy).In(img.Rect) {
				img.SetRGBA(xx, yy, c)
			}
		}
	}
}
