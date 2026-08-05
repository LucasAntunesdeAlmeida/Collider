// Memory: flip cards two at a time and find the eight pairs of pixel
// icons. Menu, drawn cards, a move counter, and a win screen. Still a
// complete game with zero movement and zero physics: clicks are
// point-vs-object collisions.
//
// Run from this folder: cd examples/memory && go run .
package main

import (
	"fmt"
	"slices"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

const font = "fonts/pixel.ttf"

func main() {
	g := engine.New("Memory", 800, 600)

	// --- Menu ---
	menu := g.Scene("menu")
	menu.Music("audios/calm.wav")
	menu.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	menu.Add(engine.Text("MEMORY").At(400, 150).Font(font).TextSize(52).TextColor(engine.Yellow))
	menu.Add(engine.Text("FIND THE EIGHT PAIRS").At(400, 225).Font(font).TextSize(14))
	menu.Add(engine.Button("PLAY").At(400, 360).Font(font).TextSize(24).Color(engine.Green)).
		OnClick(func() { g.Restart("play") })

	// --- Table ---
	play := g.Scene("play")
	play.Music("audios/calm.wav")
	play.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	hud := play.Add(engine.Text("MOVES 0").At(90, 24).Font(font).TextSize(14))

	deck := engine.Shuffle([]int{1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8})
	var flipped []*engine.Object
	matched := 0
	moves := 0

	// --- Win screen ---
	win := g.Scene("win")
	win.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	win.Add(engine.Text("PERFECT MEMORY").At(400, 200).Font(font).TextSize(28).TextColor(engine.Yellow))
	result := win.Add(engine.Text("").At(400, 280).Font(font).TextSize(16))
	win.Add(engine.Button("AGAIN").At(290, 420).Font(font).TextSize(18).Color(engine.Green)).
		OnClick(func() {
			matched, moves = 0, 0
			flipped = nil
			hud.SetText("MOVES 0")
			g.Restart("play")
		})
	win.Add(engine.Button("MENU").At(540, 420).Font(font).TextSize(18)).
		OnClick(func() { g.Go("menu") })

	for i, face := range deck {
		col, row := i%4, i/4
		card := play.Add(engine.Sprite("sprites/back.png").
			At(230+float64(col)*114, 130+float64(row)*118))
		card.Data = face

		card.OnClick(func() {
			if len(flipped) == 2 || slices.Contains(flipped, card) {
				return
			}
			g.Sound("audios/flip.wav")
			card.SetSprite(fmt.Sprintf("sprites/face%d.png", card.Data))
			flipped = append(flipped, card)

			if len(flipped) == 2 {
				moves++
				hud.SetText(fmt.Sprintf("MOVES %d", moves))
				a, b := flipped[0], flipped[1]
				play.After(0.8, func() {
					if a.Data == b.Data {
						a.Destroy()
						b.Destroy()
						g.Sound("audios/match.wav")
						matched++
						if matched == 8 {
							g.Sound("audios/win.wav")
							result.SetText(fmt.Sprintf("%d MOVES", moves))
							g.Go("win")
						}
					} else {
						a.SetSprite("sprites/back.png")
						b.SetSprite("sprites/back.png")
					}
					flipped = nil
				})
			}
		})
	}

	g.Run("menu")
}
