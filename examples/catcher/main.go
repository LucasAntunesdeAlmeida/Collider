// Catcher: catch falling gems with the basket for 30 seconds. Your
// best score is saved to save.json and greets you on the next run.
//
// This example proves a save system, and proves it needs no engine
// support: loading and saving is ~15 lines of plain encoding/json.
//
// Run from this folder: cd examples/catcher && go run .
package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

const (
	font      = "fonts/pixel.ttf"
	saveFile  = "save.json"
	roundTime = 30.0
)

// saveData is everything the game persists. Extend it and old files
// keep loading: missing fields just stay zero.
type saveData struct {
	Best int `json:"best"`
}

func load() saveData {
	var s saveData
	b, err := os.ReadFile(saveFile)
	if err != nil {
		return s // first run: zero values
	}
	json.Unmarshal(b, &s)
	return s
}

func (s saveData) write() {
	b, _ := json.MarshalIndent(s, "", "  ")
	os.WriteFile(saveFile, b, 0o644)
}

func main() {
	state := load()

	g := engine.New("Catcher", 800, 600)

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
				state.write() // the actual save, at the moment it matters
				title.SetText("RECORD")
				record.SetText("NEW BEST SAVED")
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
		gem := play.Add(engine.Sprite("sprites/gem.png").
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
