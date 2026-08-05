// Command genassets draws every example game's art and sound from
// code: pixel art as ASCII grids (sprites.go), backgrounds and tiles as
// small generators (art.go), sounds as synthesized waveforms, plus a
// copy of the Press Start 2P font (with its OFL license) into each
// game that uses text.
//
// Run from the repo root:
//
//	go run ./tools/genassets
package main

import (
	"bytes"
	"embed"
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

//go:embed fontsrc
var fontsrc embed.FS

// fontUsers get a copy of the pixel font and its license.
var fontUsers = []string{
	"agent", "arena", "breakout", "catcher", "caves", "dialog",
	"dungeon", "jumper", "memory", "pong", "runner", "ship",
	"zombie-night",
}

func main() {
	fonts()
	pong()
	breakout()
	catcher()
	jumper()
	runner()
	zombieNight()
	arena()
	gemRush()
	caves()
	dungeon()
	memory()
	dialog()
	ship()
	fmt.Println("assets generated")
}

func fonts() {
	ttf, err := fontsrc.ReadFile("fontsrc/pixel.ttf")
	if err != nil {
		panic(err)
	}
	ofl, err := fontsrc.ReadFile("fontsrc/OFL.txt")
	if err != nil {
		panic(err)
	}
	for _, game := range fontUsers {
		writeFile("examples/"+game+"/fonts/pixel.ttf", ttf)
		writeFile("examples/"+game+"/fonts/OFL.txt", ofl)
	}
}

func pong() {
	d := "examples/pong/"
	art(d+"sprites/paddle.png", 3, paddleVArt)
	art(d+"sprites/ball.png", 3, ballArt)
	starfield(d+"sprites/background.png", 800, 600, 90, rgb(0x0A0E1A), rgb(0x151C33))
	writeWav(d+"audios/bounce.wav", tone(440, 0.07, 0.5))
	writeWav(d+"audios/score.wav", seq(tone(330, 0.08, 0.4), tone(220, 0.12, 0.4)))
	writeWav(d+"audios/win.wav", melody(0.11, 0.35, 523, 659, 784, 1047))
	writeWav(d+"audios/menu.wav", melody(0.36, 0.16, 262, 330, 392, 330))
}

func breakout() {
	d := "examples/breakout/"
	art(d+"sprites/paddle.png", 4, paddleArt)
	art(d+"sprites/ball.png", 3, ballArt)
	bricks := []struct {
		name             string
		fill, light, dim rune
	}{
		{"red", 'R', 'p', 'r'},
		{"orange", 'o', 'Y', 'O'},
		{"yellow", 'Y', 'W', 'y'},
		{"green", 'G', 'h', 'g'},
		{"blue", 'B', 'c', 'b'},
	}
	for _, b := range bricks {
		art(d+"sprites/brick-"+b.name+".png", 6, brickArt(b.fill, b.light, b.dim))
	}
	backdrop(d+"sprites/background.png", 800, 600, rgb(0x101830), rgb(0x1E2A4A))
	writeWav(d+"audios/break.wav", tone(660, 0.06, 0.5))
	writeWav(d+"audios/bounce.wav", tone(392, 0.06, 0.45))
	writeWav(d+"audios/lose.wav", seq(tone(262, 0.12, 0.4), tone(196, 0.18, 0.4)))
	writeWav(d+"audios/win.wav", melody(0.11, 0.35, 523, 659, 784, 1047))
	writeWav(d+"audios/menu.wav", melody(0.3, 0.18, 262, 330, 392, 330))
	writeWav(d+"audios/action.wav", melody(0.16, 0.2, 392, 392, 440, 494, 440, 392, 587, 494))
}

func catcher() {
	d := "examples/catcher/"
	art(d+"sprites/basket.png", 4, paddleArt)
	strip(d+"sprites/gem.png", 3, gemSparkle)
	backdrop(d+"sprites/background.png", 800, 600, rgb(0x1B1030), rgb(0x33204F))
	writeWav(d+"audios/catch.wav", seq(tone(880, 0.04, 0.35), tone(1175, 0.06, 0.35)))
	writeWav(d+"audios/miss.wav", tone(196, 0.1, 0.3))
	writeWav(d+"audios/end.wav", melody(0.13, 0.3, 784, 659, 523))
	writeWav(d+"audios/theme.wav", melody(0.24, 0.16, 330, 392, 494, 392))
}

func jumper() {
	d := "examples/jumper/"
	strip(d+"sprites/player.png", 3, heroWalk)
	art(d+"sprites/idle.png", 3, heroWalk[0])
	strip(d+"sprites/coin.png", 3, coinSpin)
	art(d+"sprites/spikes.png", 3, spikeArt)
	tile(d+"sprites/platform.png", 40, 20, rgb(0x3EB658), rgb(0x7FE39A), rgb(0x2C7A46), 4)
	tile(d+"sprites/ground.png", 40, 40, rgb(0x6B4423), rgb(0x8B5A2B), rgb(0x4A2F18), 6)
	hills(d+"sprites/background.png", 800, 600, rgb(0x4FA3E0), rgb(0x2C7A46))
	writeWav(d+"audios/jump.wav", seq(tone(392, 0.04, 0.35), tone(587, 0.06, 0.35)))
	writeWav(d+"audios/coin.wav", seq(tone(880, 0.05, 0.4), tone(1320, 0.08, 0.4)))
	writeWav(d+"audios/hurt.wav", seq(tone(220, 0.1, 0.45), noise(0.1, 0.3)))
	writeWav(d+"audios/win.wav", melody(0.12, 0.35, 523, 659, 784, 1047))
	writeWav(d+"audios/level1.wav", melody(0.22, 0.22, 392, 440, 494, 587, 494, 440, 392, 330))
	writeWav(d+"audios/menu.wav", melody(0.34, 0.18, 330, 392, 494, 392))
}

func runner() {
	d := "examples/runner/"
	strip(d+"sprites/run.png", 3, runnerRun)
	strip(d+"sprites/jump.png", 3, runnerJump)
	strip(d+"sprites/death.png", 3, runnerDeath)
	art(d+"sprites/crate.png", 3, crateArt)
	tile(d+"sprites/ground.png", 40, 40, rgb(0x8B5A2B), rgb(0xB07C4A), rgb(0x6B4423), 8)
	hills(d+"sprites/background.png", 800, 600, rgb(0xE08A4F), rgb(0x6B4423))
	writeWav(d+"audios/jump.wav", seq(tone(330, 0.05, 0.35), tone(494, 0.07, 0.35)))
	writeWav(d+"audios/hit.wav", noise(0.16, 0.45))
	writeWav(d+"audios/run.wav", melody(0.14, 0.18, 330, 392, 440, 392, 494, 440, 392, 330))
	writeWav(d+"audios/menu.wav", melody(0.3, 0.18, 262, 330, 392, 330))
}

func zombieNight() {
	d := "examples/zombie-night/"
	strip(d+"sprites/hero.png", 3, heroWalk)
	art(d+"sprites/hero-idle.png", 3, heroWalk[0])
	strip(d+"sprites/zombie.png", 3, zombieWalk)
	art(d+"sprites/bullet.png", 3, bulletArt)
	tile(d+"sprites/ground.png", 40, 40, rgb(0x1B2418), rgb(0x2A3524), rgb(0x121A10), 10)
	backdrop(d+"sprites/background.png", 800, 600, rgb(0x0C1208), rgb(0x1B2418))
	writeWav(d+"audios/shot.wav", noise(0.09, 0.4))
	writeWav(d+"audios/kill.wav", seq(noise(0.06, 0.3), tone(147, 0.1, 0.3)))
	writeWav(d+"audios/death.wav", seq(tone(147, 0.16, 0.45), noise(0.2, 0.3)))
	writeWav(d+"audios/dread.wav", melody(0.5, 0.16, 110, 104, 110, 98))
	writeWav(d+"audios/menu.wav", melody(0.45, 0.16, 147, 139, 131, 139))
}

func arena() {
	d := "examples/arena/"
	strip(d+"sprites/player.png", 3, heroWalk)
	art(d+"sprites/idle.png", 3, heroWalk[0])
	strip(d+"sprites/chaser.png", 3, zombieWalk)
	art(d+"sprites/heart.png", 3, []string{
		" RR  RR ",
		"RRRRRRRR",
		"RRRRRRRR",
		" RRRRRR ",
		"  RRRR  ",
		"   RR   ",
	})
	tile(d+"sprites/floor.png", 40, 40, rgb(0x33204F), rgb(0x4A3070), rgb(0x231438), 6)
	backdrop(d+"sprites/background.png", 800, 600, rgb(0x1B1030), rgb(0x33204F))
	writeWav(d+"audios/hurt.wav", seq(tone(196, 0.1, 0.45), noise(0.08, 0.3)))
	writeWav(d+"audios/death.wav", melody(0.16, 0.35, 392, 330, 262, 196))
	writeWav(d+"audios/theme.wav", melody(0.26, 0.16, 220, 262, 330, 262))
}

func gemRush() {
	d := "examples/agent/"
	strip(d+"sprites/player.png", 3, heroWalk)
	art(d+"sprites/idle.png", 3, heroWalk[0])
	strip(d+"sprites/gem.png", 3, gemSparkle)
	tile(d+"sprites/floor.png", 40, 40, rgb(0x1E2A4A), rgb(0x2C3C66), rgb(0x141C33), 5)
	backdrop(d+"sprites/background.png", 800, 600, rgb(0x101830), rgb(0x1E2A4A))
	writeWav(d+"audios/gem.wav", seq(tone(1047, 0.04, 0.35), tone(1568, 0.07, 0.35)))
	writeWav(d+"audios/win.wav", melody(0.12, 0.35, 523, 659, 784, 1047))
	writeWav(d+"audios/lose.wav", melody(0.16, 0.3, 392, 311, 262))
}

func caves() {
	d := "examples/caves/"
	strip(d+"sprites/player.png", 3, heroWalk)
	strip(d+"sprites/gem.png", 2, gemSparkle)
	tile(d+"sprites/rock.png", 20, 20, rgb(0x4A4238), rgb(0x6B5F50), rgb(0x2E2822), 4)
	backdrop(d+"sprites/background.png", 800, 600, rgb(0x120E0A), rgb(0x241C14))
	writeWav(d+"audios/gem.wav", seq(tone(1047, 0.05, 0.35), tone(1568, 0.08, 0.35)))
	writeWav(d+"audios/win.wav", melody(0.12, 0.35, 523, 659, 784, 1047))
	writeWav(d+"audios/theme.wav", melody(0.5, 0.14, 196, 220, 165, 220))
}

func dungeon() {
	d := "examples/dungeon/"
	strip(d+"sprites/player.png", 3, heroWalk)
	art(d+"sprites/key.png", 3, keyArt)
	art(d+"sprites/door.png", 3, doorArt)
	art(d+"sprites/chest.png", 3, chestArt)
	tile(d+"sprites/wall.png", 40, 40, rgb(0x5A5348), rgb(0x7D7466), rgb(0x3A352E), 5)
	tile(d+"sprites/floor.png", 40, 40, rgb(0x2A2620), rgb(0x3A342C), rgb(0x1C1915), 4)
	backdrop(d+"sprites/background.png", 800, 600, rgb(0x14110D), rgb(0x241F18))
	writeWav(d+"audios/key.wav", seq(tone(784, 0.05, 0.35), tone(1047, 0.08, 0.35)))
	writeWav(d+"audios/door.wav", seq(tone(196, 0.12, 0.4), noise(0.08, 0.2)))
	writeWav(d+"audios/win.wav", melody(0.12, 0.35, 523, 659, 784, 1047))
	writeWav(d+"audios/dungeon.wav", melody(0.45, 0.14, 147, 165, 131, 165))
}

func memory() {
	d := "examples/memory/"
	art(d+"sprites/back.png", 8, cardBack)
	for i, face := range cardFaces {
		art(fmt.Sprintf("%ssprites/face%d.png", d, i+1), 8, face)
	}
	backdrop(d+"sprites/background.png", 800, 600, rgb(0x14301E), rgb(0x1F4A2E))
	writeWav(d+"audios/flip.wav", tone(587, 0.04, 0.3))
	writeWav(d+"audios/match.wav", seq(tone(523, 0.06, 0.4), tone(784, 0.1, 0.4)))
	writeWav(d+"audios/win.wav", melody(0.12, 0.35, 523, 659, 784, 1047))
	writeWav(d+"audios/calm.wav", melody(0.6, 0.14, 262, 294, 330, 294))
}

func dialog() {
	d := "examples/dialog/"
	art(d+"sprites/elder.png", 6, elderArt)
	art(d+"sprites/hero.png", 6, heroArt)
	backdrop(d+"sprites/background.png", 800, 600, rgb(0x1A1428), rgb(0x2E2440))
	writeWav(d+"audios/blip.wav", tone(880, 0.02, 0.18))
	writeWav(d+"audios/theme.wav", melody(0.5, 0.14, 294, 330, 392, 330, 294, 262))
}

func ship() {
	d := "examples/ship/"
	strip(d+"sprites/ship.png", 3, shipFly)
	strip(d+"sprites/boom.png", 3, boomFrames)
	art(d+"sprites/comet.png", 3, cometArt)
	art(d+"sprites/icon.png", 2, shipFly[0])
	starfield(d+"sprites/background.png", 800, 600, 170, rgb(0x0A0E1A), rgb(0x141B2E))
	writeWav(d+"audios/hit.wav", noise(0.18, 0.45))
	writeWav(d+"audios/theme.wav", melody(0.2, 0.18, 262, 330, 392, 523, 392, 330))
}

// cometArt is the falling fireball in the ship game.
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

// shipFly is the player ship with its thruster flame, two frames.
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
		"     o    o     ",
		"    oo    oo    ",
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
		"    oo    oo    ",
		"    Yo    oY    ",
		"    YY    YY    ",
	},
}

// --- file writing ---

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

func writeFile(path string, data []byte) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		panic(err)
	}
	fmt.Println("  ", path)
}

// --- sound synthesis ---

// writeWav writes 16-bit mono PCM at 48kHz.
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

// noise is a white-noise burst with decay, for shots and impacts.
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

var _ color.RGBA // art.go owns the color helpers
