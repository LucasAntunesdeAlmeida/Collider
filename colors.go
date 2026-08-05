package collider

import "image/color"

// Color is any standard library color. The palette below covers the
// common cases so examples never need to import image/color.
type Color = color.Color

var (
	White  = color.RGBA{R: 0xF2, G: 0xF2, B: 0xF2, A: 0xFF}
	Black  = color.RGBA{R: 0x10, G: 0x10, B: 0x10, A: 0xFF}
	Red    = color.RGBA{R: 0xE5, G: 0x3E, B: 0x3E, A: 0xFF}
	Green  = color.RGBA{R: 0x3E, G: 0xB6, B: 0x58, A: 0xFF}
	Blue   = color.RGBA{R: 0x3E, G: 0x6F, B: 0xE5, A: 0xFF}
	Yellow = color.RGBA{R: 0xF2, G: 0xC9, B: 0x38, A: 0xFF}
	Orange = color.RGBA{R: 0xF2, G: 0x8C, B: 0x38, A: 0xFF}
)
