// Catcher: catch falling gems with the paddle for 30 seconds. Your best
// score survives the program: it is saved to save.json and loaded on
// the next run.
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

// saveData is everything the game persists. Extend it and old files
// keep loading: missing fields just stay zero.
type saveData struct {
	Best int `json:"best"`
}

const saveFile = "save.json"

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
	play := g.Scene("play")

	paddle := play.Add(engine.Rect(120, 20, engine.White).At(400, 560).Solid())
	hud := play.Add(engine.Text(fmt.Sprintf("Score: 0   Best: %d", state.Best)).At(400, 24).TextSize(20))

	score := 0
	paddle.OnUpdate(func(_ float64) {
		x, _ := g.Mouse()
		paddle.X = x
	})

	play.Every(0.5, func() {
		gem := play.Add(engine.Rect(14, 14, engine.Yellow).
			At(20+rand.Float64()*760, -10).Tag("gem"))
		gem.Vy = 220 + rand.Float64()*180
		gem.LifeTime(4)
	})

	paddle.OnCollisionWith("gem", func(gem *engine.Object) {
		gem.Destroy()
		score++
		g.Sound("audios/catch.wav")
		hud.SetText(fmt.Sprintf("Score: %d   Best: %d", score, state.Best))
	})

	end := g.Scene("end")
	result := end.Add(engine.Text("").At(400, 270).TextSize(30))
	end.Add(engine.Text("Click to play again.").At(400, 340))
	end.OnClick(func(_, _ float64) {
		score = 0
		hud.SetText(fmt.Sprintf("Score: 0   Best: %d", state.Best))
		g.Restart("play")
	})

	play.After(30, func() {
		msg := fmt.Sprintf("Time! You caught %d gems.", score)
		if score > state.Best {
			state.Best = score
			state.write() // the actual save: one call at the moment it matters
			msg = fmt.Sprintf("New record: %d gems!", score)
		}
		result.SetText(msg)
		g.Go("end")
	})

	g.Run("play")
}
