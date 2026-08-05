// Package game is the shared game definition for the agent example:
// collect all gems before the time runs out. Built to be played by
// humans (main.go) and by programs or AI agents (bot/, or MCP via
// COLLIDER_AGENT=mcp), which is why every object that matters carries
// a tag and the HUD is a text object agents can read.
package game

import (
	"fmt"
	"math/rand/v2"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

const (
	Gems      = 6
	TimeLimit = 30.0
)

// New builds the game. The caller decides how it runs: g.Run for a
// window, g.Headless plus g.Step for a bot.
func New() *engine.Game {
	g := engine.New("Gem Rush", 800, 600)
	play := g.Scene("play")

	player := play.Add(engine.Rect(24, 24, engine.Blue).At(400, 300).Tag("player"))
	hud := play.Add(engine.Text("").At(110, 24).TextSize(18).Tag("hud"))

	for range Gems {
		play.Add(engine.Rect(16, 16, engine.Yellow).
			At(40+rand.Float64()*720, 40+rand.Float64()*520).Tag("gem"))
	}

	left := TimeLimit
	player.OnUpdate(func(dt float64) {
		if g.Key(engine.W) || g.Key(engine.Up) {
			player.Move(0, -240*dt)
		}
		if g.Key(engine.S) || g.Key(engine.Down) {
			player.Move(0, 240*dt)
		}
		if g.Key(engine.A) || g.Key(engine.Left) {
			player.Move(-240*dt, 0)
		}
		if g.Key(engine.D) || g.Key(engine.Right) {
			player.Move(240*dt, 0)
		}
		left -= dt
		hud.SetText(fmt.Sprintf("gems left %d  time %.0f", play.Count("gem"), left))
		if left <= 0 {
			g.Go("lose")
		}
	})

	player.OnCollisionWith("gem", func(gem *engine.Object) {
		gem.Destroy()
		if play.Count("gem") == 0 {
			g.Go("win")
		}
	})

	win := g.Scene("win")
	win.Add(engine.Text("All gems collected!").At(400, 300).Tag("result"))

	lose := g.Scene("lose")
	lose.Add(engine.Text("Time's up.").At(400, 300).Tag("result"))

	return g
}
