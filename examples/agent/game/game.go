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

	font = "fonts/pixel.ttf"
)

// New builds the game. The caller decides how it runs: g.Run for a
// window, g.Headless plus g.Step for a bot.
func New() *engine.Game {
	g := engine.New("Gem Rush", 800, 600)

	// --- Menu (humans start here; bots jump straight to "play") ---
	menu := g.Scene("menu")
	menu.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	menu.Add(engine.Text("GEM RUSH").At(400, 150).Font(font).TextSize(44).TextColor(engine.Yellow))
	menu.Add(engine.Text("6 GEMS   30 SECONDS").At(400, 225).Font(font).TextSize(14))
	menu.Add(engine.Text("ALSO PLAYABLE BY BOTS AND AI AGENTS").At(400, 500).Font(font).TextSize(10))
	menu.Add(engine.Button("PLAY").At(400, 360).Font(font).TextSize(24).Color(engine.Green)).
		OnClick(func() { g.Restart("play") })

	// --- Round ---
	play := g.Scene("play")
	play.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())

	player := play.Add(engine.Rect(32, 45, nil).At(400, 300).Tag("player").
		Animation("walk", "sprites/player.png", 2, 8).
		Animation("idle", "sprites/idle.png", 1, 1))
	player.Play("idle")

	hud := play.Add(engine.Text("").At(140, 24).Font(font).TextSize(14).Tag("hud"))

	for range Gems {
		gem := play.Add(engine.Rect(30, 24, nil).
			At(40+rand.Float64()*720, 60+rand.Float64()*500).Tag("gem").
			Animation("sparkle", "sprites/gem.png", 2, 4))
		gem.Play("sparkle")
	}

	left := TimeLimit
	player.OnUpdate(func(dt float64) {
		moving := false
		if g.Key(engine.W) || g.Key(engine.Up) {
			player.Move(0, -240*dt)
			moving = true
		}
		if g.Key(engine.S) || g.Key(engine.Down) {
			player.Move(0, 240*dt)
			moving = true
		}
		if g.Key(engine.A) || g.Key(engine.Left) {
			player.Move(-240*dt, 0)
			moving = true
		}
		if g.Key(engine.D) || g.Key(engine.Right) {
			player.Move(240*dt, 0)
			moving = true
		}
		if moving {
			player.Play("walk")
		} else {
			player.Play("idle")
		}
		left -= dt
		hud.SetText(fmt.Sprintf("GEMS %d  TIME %.0f", play.Count("gem"), left))
		if left <= 0 {
			left = TimeLimit
			g.Sound("audios/lose.wav")
			g.Go("lose")
		}
	})

	player.OnCollisionWith("gem", func(gem *engine.Object) {
		gem.Destroy()
		g.Sound("audios/gem.wav")
		if play.Count("gem") == 0 {
			left = TimeLimit
			g.Sound("audios/win.wav")
			g.Go("win")
		}
	})

	// --- End screens ---
	win := g.Scene("win")
	win.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	win.Add(engine.Text("ALL GEMS COLLECTED").At(400, 240).Font(font).TextSize(24).
		TextColor(engine.Yellow).Tag("result"))
	win.Add(engine.Button("AGAIN").At(400, 400).Font(font).TextSize(18).Color(engine.Green)).
		OnClick(func() { g.Restart("play") })

	lose := g.Scene("lose")
	lose.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	lose.Add(engine.Text("TIME'S UP").At(400, 240).Font(font).TextSize(28).
		TextColor(engine.Orange).Tag("result"))
	lose.Add(engine.Button("AGAIN").At(400, 400).Font(font).TextSize(18).Color(engine.Green)).
		OnClick(func() { g.Restart("play") })

	return g
}
