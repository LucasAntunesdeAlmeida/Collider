// Ship: a small dodge game built to be PUBLISHED. Pixel-art sprites,
// an animated thruster, an explosion, a starfield, a real menu with
// drawn buttons, and the Press Start 2P pixel font (OFL licensed, see
// fonts/OFL.txt). Assets are embedded: the binary ships alone.
//
// A/D or arrows to dodge. F toggles fullscreen.
//
// Run: cd examples/ship && go run .
// Ship: see PUBLISHING.md at the repo root.
package main

import (
	"embed"
	"fmt"
	"math/rand/v2"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

//go:embed sprites audios fonts
var content embed.FS

const font = "fonts/pixel.ttf"

func main() {
	engine.UseAssets(content)

	g := engine.New("Ship", 800, 600)
	g.Resizable(true)
	g.Icon("sprites/icon.png")

	best := 0.0

	// F toggles fullscreen in every scene.
	isFull := false
	fsDown := false
	fullscreenKey := func(_ float64) {
		if g.Key(engine.F) && !fsDown {
			isFull = !isFull
			g.Fullscreen(isFull)
		}
		fsDown = g.Key(engine.F)
	}

	// --- Menu ---
	menu := g.Scene("menu")
	menu.Music("audios/theme.wav")
	menu.Add(engine.Sprite("sprites/background.png").At(400, 300))
	menu.Add(engine.Text("SHIP").At(400, 160).Font(font).TextSize(64).TextColor(engine.Yellow))
	menu.Add(engine.Text("DODGE THE COMETS").At(400, 240).Font(font).TextSize(16))
	playBtn := menu.Add(engine.Sprite("sprites/button.png").At(400, 350))
	menu.Add(engine.Text("PLAY").At(400, 350).Font(font).TextSize(20).TextColor(engine.Yellow))
	menu.Add(engine.Text("A/D MOVE   F FULLSCREEN").At(400, 520).Font(font).TextSize(12))
	playBtn.OnClick(func() { g.Go("play") })
	menu.OnUpdate(fullscreenKey)

	// --- Gameplay ---
	play := g.Scene("play")
	play.Music("audios/theme.wav")
	play.OnUpdate(fullscreenKey)
	play.Add(engine.Sprite("sprites/background.png").At(400, 300))

	ship := play.Add(engine.Rect(48, 48, engine.Blue).At(400, 520).
		Animation("fly", "sprites/ship.png", 2, 8).
		Animation("boom", "sprites/boom.png", 4, 8))
	ship.Play("fly")

	hud := play.Add(engine.Text("0.0").At(80, 30).Font(font).TextSize(16))

	alive := true
	timeAlive := 0.0

	ship.OnUpdate(func(dt float64) {
		if !alive {
			return
		}
		timeAlive += dt
		hud.SetText(fmt.Sprintf("%.1f", timeAlive))
		if g.Key(engine.A) || g.Key(engine.Left) {
			ship.Move(-340*dt, 0)
		}
		if g.Key(engine.D) || g.Key(engine.Right) {
			ship.Move(340*dt, 0)
		}
		if ship.X < 24 {
			ship.X = 24
		}
		if ship.X > 776 {
			ship.X = 776
		}
	})

	play.Every(0.4, func() {
		if !alive {
			return
		}
		c := play.Add(engine.Sprite("sprites/comet.png").
			At(20+rand.Float64()*760, -30).Tag("comet"))
		c.Vy = 240 + rand.Float64()*(140+timeAlive*6)
		c.Vx = (rand.Float64() - 0.5) * 80
		c.LifeTime(5)
	})

	// --- Game over ---
	over := g.Scene("gameover")
	over.Music("audios/theme.wav")
	over.OnUpdate(fullscreenKey)
	over.Add(engine.Sprite("sprites/background.png").At(400, 300))
	over.Add(engine.Sprite("sprites/panel.png").At(400, 290))
	over.Add(engine.Text("CRASHED").At(400, 240).Font(font).TextSize(28).TextColor(engine.Orange))
	result := over.Add(engine.Text("").At(400, 295).Font(font).TextSize(14))
	bestText := over.Add(engine.Text("").At(400, 330).Font(font).TextSize(14).TextColor(engine.Yellow))
	againBtn := over.Add(engine.Sprite("sprites/button.png").At(290, 430))
	over.Add(engine.Text("AGAIN").At(290, 430).Font(font).TextSize(16).TextColor(engine.Yellow))
	menuBtn := over.Add(engine.Sprite("sprites/button.png").At(510, 430))
	over.Add(engine.Text("MENU").At(510, 430).Font(font).TextSize(16).TextColor(engine.Yellow))

	reset := func(target string) {
		alive = true
		timeAlive = 0
		ship.Play("fly")
		g.Restart("play")
		if target != "play" {
			g.Go(target)
		}
	}
	againBtn.OnClick(func() { reset("play") })
	menuBtn.OnClick(func() { reset("menu") })

	ship.OnCollisionWith("comet", func(_ *engine.Object) {
		if !alive {
			return
		}
		alive = false
		ship.PlayOnce("boom")
		g.Sound("audios/hit.wav")
		if timeAlive > best {
			best = timeAlive
		}
		result.SetText(fmt.Sprintf("YOU SURVIVED %.1f S", timeAlive))
		bestText.SetText(fmt.Sprintf("BEST %.1f S", best))
		play.After(0.9, func() { g.Go("gameover") })
	})

	g.Run("menu")
}
