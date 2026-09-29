// Catcher: catch falling gems with the basket for 30 seconds. Your
// best score is saved and greets you on the next run, on desktop and
// in the browser alike.
//
// This example proves the save system: g.Load at startup, g.Save at
// the moment it matters. Values are plain JSON-encodable Go values.
// Closing the window mid-round (X, Alt+F4) still keeps a new best:
// g.OnClose saves it before the game ends.
//
// Run from this folder: cd examples/catcher && go run .
package main

import (
	"fmt"
	"math/rand/v2"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

const (
	font      = "fonts/pixel.ttf"
	roundTime = 30.0
)

// saveData is everything the game persists. Extend it and old saves
// keep loading: missing fields keep the values they had before Load.
type saveData struct {
	Best int `json:"best"`
}

func main() {
	g := engine.New("Catcher", 800, 600)

	var state saveData
	g.Load("save", &state) // false on the first run: state stays zero

	// --- Menu ---
	menu := g.Scene("menu")
	menu.Music("audios/theme.wav")
	menu.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	menu.Add(engine.Text("CATCHER").At(400, 150).Font(font).TextSize(52).TextColor(engine.Yellow))
	menu.Add(engine.Text("30 SECONDS OF FALLING GEMS").At(400, 225).Font(font).TextSize(12))
	bestLabel := menu.Add(engine.Text("").At(400, 275).Font(font).TextSize(16).TextColor(engine.Green))
	menu.Add(engine.Text("MOUSE MOVES THE BASKET").At(400, 500).Font(font).TextSize(12))
	menu.Add(engine.Button("PLAY").At(400, 380).Font(font).TextSize(24).Color(engine.Green)).
		OnClick(func() { g.Restart("play") })

	showBest := func() {
		if state.Best > 0 {
			bestLabel.SetText(fmt.Sprintf("BEST %d GEMS", state.Best))
		}
	}
	showBest()

	// --- Game ---
	play := g.Scene("play")
	play.Music("audios/theme.wav")
	play.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())

	basket := play.Add(engine.Sprite("sprites/basket.png").At(400, 545).Solid().Tag("basket"))
	hud := play.Add(engine.Text("").At(150, 30).Font(font).TextSize(16))

	score := 0
	left := roundTime
	updateHUD := func() {
		hud.SetText(fmt.Sprintf("GEMS %d   TIME %02.0f", score, left))
	}
	updateHUD()

	// Closing the window mid-round would lose a record in the making:
	// OnClose runs once, on the game loop, before the game ends.
	g.OnClose(func() {
		if score > state.Best { // score is 0 outside a round
			state.Best = score
			g.Save("save", state)
		}
	})

	// --- End screen ---
	over := g.Scene("over")
	over.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	title := over.Add(engine.Text("TIME").At(400, 200).Font(font).TextSize(44).TextColor(engine.Yellow))
	result := over.Add(engine.Text("").At(400, 285).Font(font).TextSize(18))
	record := over.Add(engine.Text("").At(400, 335).Font(font).TextSize(14).TextColor(engine.Green))
	over.Add(engine.Button("AGAIN").At(290, 440).Font(font).TextSize(18).Color(engine.Green)).
		OnClick(func() { g.Restart("play") })
	over.Add(engine.Button("MENU").At(540, 440).Font(font).TextSize(18)).
		OnClick(func() { g.Go("menu") })

	basket.OnUpdate(func(dt float64) {
		x, _ := g.Mouse()
		basket.X = clamp(x, 60, 740)

		left -= dt
		updateHUD()
		if left <= 0 {
			result.SetText(fmt.Sprintf("YOU CAUGHT %d GEMS", score))
			if score > state.Best {
				state.Best = score
				title.SetText("RECORD")
				record.SetText("NEW BEST SAVED")
				// The actual save, at the moment it matters.
				if err := g.Save("save", state); err != nil {
					record.SetText("NEW BEST (NOT SAVED)")
				}
			} else {
				title.SetText("TIME")
				record.SetText(fmt.Sprintf("BEST %d", state.Best))
			}
			showBest()
			g.Sound("audios/end.wav")
			score, left = 0, roundTime
			updateHUD()
			g.Go("over")
		}
	})

	play.Every(0.5, func() {
		// Rect(nil) + Animation = an object drawn only by its frames.
		gem := play.Add(engine.Rect(30, 24, nil).
			At(30+rand.Float64()*740, -20).Tag("gem").
			Animation("spin", "sprites/gem.png", 2, 6))
		gem.Play("spin")
		gem.Vy = 200 + rand.Float64()*200
		gem.LifeTime(5)
	})

	basket.OnCollisionWith("gem", func(gem *engine.Object) {
		gem.Destroy()
		score++
		updateHUD()
		g.Sound("audios/catch.wav")
	})

	g.Run("menu")
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
