// Memory: the counter-example, a complete game with no movement and no
// physics at all. Clicks are point-vs-object collisions, so a click-driven
// game is still squarely inside the engine's model.
//
// Run from this folder so the assets resolve: cd examples/memory && go run .
package main

import (
	"fmt"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

func main() {
	g := engine.New("Memory", 800, 600)
	play := g.Scene("play")
	play.Music("calm.wav")

	// 8 pairs, shuffled onto a 4x4 grid.
	deck := engine.Shuffle([]int{1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8})
	var flipped []*engine.Object
	matched := 0

	for i, face := range deck {
		col, row := i%4, i/4
		card := play.Add(engine.Sprite("back.png").
			At(220+float64(col)*120, 120+float64(row)*120))
		card.Data = face

		card.OnClick(func() {
			if len(flipped) == 2 || card.Data == nil {
				return
			}
			card.SetSprite(fmt.Sprintf("face%d.png", card.Data))
			flipped = append(flipped, card)

			if len(flipped) == 2 {
				a, b := flipped[0], flipped[1]
				play.After(0.8, func() {
					if a.Data == b.Data {
						a.Destroy()
						b.Destroy()
						g.Sound("match.wav")
						matched++
						if matched == 8 {
							g.Go("win")
						}
					} else {
						a.SetSprite("back.png")
						b.SetSprite("back.png")
					}
					flipped = nil
				})
			}
		})
	}

	win := g.Scene("win")
	win.Add(engine.Text("You remembered everything!").At(400, 300))
	win.OnClick(func(_, _ float64) {
		matched = 0
		flipped = nil
		g.Restart("play")
	})

	g.Run("play")
}
