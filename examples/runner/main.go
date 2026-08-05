// Runner: an endless runner with a drawn sprinter. The world moves,
// the player stays; Space is the only input. Menu, run/jump/death
// animations, tiled ground, sunset hills.
//
// Run from this folder: cd examples/runner && go run .
package main

import (
	"fmt"
	"math/rand/v2"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

const font = "fonts/pixel.ttf"

func main() {
	g := engine.New("Runner", 800, 600)

	// --- Menu ---
	menu := g.Scene("menu")
	menu.Music("audios/menu.wav")
	menu.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	menu.Add(engine.Text("RUNNER").At(400, 150).Font(font).TextSize(52).TextColor(engine.Yellow))
	menu.Add(engine.Text("JUMP THE CRATES").At(400, 225).Font(font).TextSize(14))
	menu.Add(engine.Text("SPACE TO JUMP").At(400, 500).Font(font).TextSize(12))
	menu.Add(engine.Button("RUN").At(400, 360).Font(font).TextSize(24).Color(engine.Green)).
		OnClick(func() { g.Restart("play") })

	// --- Run ---
	play := g.Scene("play")
	play.Music("audios/run.wav")
	play.Gravity(2000)
	play.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())

	play.Add(engine.Rect(800, 60, nil).At(400, 580).Solid())
	for i := range 20 {
		play.Add(engine.Sprite("sprites/ground.png").At(float64(i)*40+20, 570).Visual())
	}

	player := play.Add(engine.Rect(36, 40, nil).At(150, 500).WithGravity().
		Animation("run", "sprites/run.png", 4, 12).
		Animation("jump", "sprites/jump.png", 2, 6).
		Animation("death", "sprites/death.png", 4, 8))
	player.Play("run")

	hud := play.Add(engine.Text("0 M").At(70, 26).Font(font).TextSize(16))
	alive := true
	dist := 0.0

	// --- Game over ---
	over := g.Scene("over")
	over.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	over.Add(engine.Text("TRIPPED!").At(400, 200).Font(font).TextSize(40).TextColor(engine.Orange))
	result := over.Add(engine.Text("").At(400, 280).Font(font).TextSize(18))
	over.Add(engine.Button("AGAIN").At(290, 420).Font(font).TextSize(18).Color(engine.Green)).
		OnClick(func() {
			alive = true
			dist = 0
			player.Play("run")
			g.Restart("play")
		})
	over.Add(engine.Button("MENU").At(540, 420).Font(font).TextSize(18)).
		OnClick(func() {
			alive = true
			dist = 0
			player.Play("run")
			g.Restart("play")
			g.Go("menu")
		})

	player.OnUpdate(func(dt float64) {
		if !alive {
			return
		}
		dist += 60 * dt
		hud.SetText(fmt.Sprintf("%d M", int(dist)))
		if g.Key(engine.Space) && player.Grounded() {
			player.Vy = -840
			g.Sound("audios/jump.wav")
		}
		if player.Grounded() {
			player.Play("run")
		} else {
			player.Play("jump")
		}
	})

	spawnCrate := func() {
		if !alive {
			return
		}
		size := 30 + rand.Float64()*24
		crate := play.Add(engine.Sprite("sprites/crate.png").Size(size, size).
			At(830, 550-size/2).Tag("crate"))
		crate.Vx = -(380 + dist/2.5) // the world speeds up as you go
		crate.LifeTime(4)
	}
	play.After(0.2, spawnCrate) // no dead air: the first crate is already coming
	play.Every(0.8, spawnCrate)

	player.OnCollisionWith("crate", func(_ *engine.Object) {
		if !alive {
			return
		}
		alive = false
		player.PlayOnce("death")
		g.Sound("audios/hit.wav")
		result.SetText(fmt.Sprintf("YOU RAN %d M", int(dist)))
		play.After(0.9, func() { g.Go("over") })
	})

	g.Run("menu")
}
