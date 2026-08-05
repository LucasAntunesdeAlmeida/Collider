// Package game is Gem Rush: collect all gems before the timer runs
// out. This game accepts NO human input by design: it exists to be
// played by agents (the built-in Decide pilot, a headless bot, or any
// MCP client). Every object that matters carries a tag and the HUD is
// a text object, so the whole game state is observable.
package game

import (
	"fmt"
	"math"
	"math/rand/v2"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

const (
	Gems      = 6
	TimeLimit = 30.0

	font = "fonts/pixel.ttf"
)

// New builds the game. The caller decides how it runs: Autopilot for a
// watchable window, Headless plus Step for a bot, MCP for external
// agents.
func New() *engine.Game {
	g := engine.New("Gem Rush", 800, 600)

	play := g.Scene("play")
	play.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	play.Add(engine.Text("AGENT AT PLAY  NO HUMAN INPUT").At(400, 580).
		Font(font).TextSize(10).Visual())

	player := play.Add(engine.Rect(32, 45, nil).At(400, 300).Tag("player").
		Animation("walk", "sprites/player.png", 2, 8).
		Animation("idle", "sprites/idle.png", 1, 1))
	player.Play("idle")

	hud := play.Add(engine.Text("").At(140, 24).Font(font).TextSize(14).Tag("hud"))

	for range Gems {
		gem := play.Add(engine.Rect(30, 24, nil).
			At(40+rand.Float64()*720, 60+rand.Float64()*480).Tag("gem").
			Animation("sparkle", "sprites/gem.png", 2, 4))
		gem.Play("sparkle")
	}

	left := TimeLimit
	player.OnUpdate(func(dt float64) {
		moving := false
		if g.Key(engine.W) {
			player.Move(0, -240*dt)
			moving = true
		}
		if g.Key(engine.S) {
			player.Move(0, 240*dt)
			moving = true
		}
		if g.Key(engine.A) {
			player.Move(-240*dt, 0)
			moving = true
		}
		if g.Key(engine.D) {
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

	// End screens auto-restart: an agent-only game keeps itself looping.
	endScene := func(name, title string, c engine.Color) {
		s := g.Scene(name)
		s.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
		s.Add(engine.Text(title).At(400, 260).Font(font).TextSize(26).
			TextColor(c).Tag("result"))
		s.Add(engine.Text("NEXT ROUND SOON").At(400, 330).Font(font).TextSize(12))
		wait := 0.0
		s.OnUpdate(func(dt float64) {
			wait += dt
			if wait > 1.5 {
				wait = 0
				g.Restart("play")
			}
		})
	}
	endScene("win", "ALL GEMS COLLECTED", engine.Yellow)
	endScene("lose", "TIME'S UP", engine.Orange)

	return g
}

// Decide is the built-in pilot: given an observation, chase the nearest
// gem. It uses only what any external agent gets: tags and positions.
func Decide(obs engine.Observation) engine.Action {
	var player *engine.ObjectObs
	var gems []engine.ObjectObs
	for i, o := range obs.Objects {
		switch o.Tag {
		case "player":
			player = &obs.Objects[i]
		case "gem":
			gems = append(gems, o)
		}
	}
	if player == nil || len(gems) == 0 {
		return engine.Action{}
	}

	target := gems[0]
	best := math.Hypot(target.X-player.X, target.Y-player.Y)
	for _, gem := range gems[1:] {
		if d := math.Hypot(gem.X-player.X, gem.Y-player.Y); d < best {
			best, target = d, gem
		}
	}

	var keys []engine.Key
	if target.X < player.X-4 {
		keys = append(keys, engine.A)
	}
	if target.X > player.X+4 {
		keys = append(keys, engine.D)
	}
	if target.Y < player.Y-4 {
		keys = append(keys, engine.W)
	}
	if target.Y > player.Y+4 {
		keys = append(keys, engine.S)
	}
	return engine.Action{Keys: keys}
}
