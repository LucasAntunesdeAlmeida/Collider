// Runner: an endless runner. The player runs automatically (the world
// moves, the player stays), Space is the only input. Survive the crates.
//
// This example proves sprite sheet animation: the player has run, jump
// and death strips, switched with Play and PlayOnce. Auto-scrolling
// needs no camera: obstacles simply move left with a velocity while the
// player's x stays fixed, the classic moving-world approach.
//
// Run from this folder: cd examples/runner && go run .
package main

import (
	"fmt"
	"math/rand/v2"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

func main() {
	g := engine.New("Runner", 800, 600)
	play := g.Scene("play")
	play.Music("audios/run.wav")
	play.Gravity(2000)

	play.Add(engine.Rect(800, 60, engine.Green).At(400, 570).Solid())

	player := play.Add(engine.Rect(32, 48, engine.Blue).At(150, 480).WithGravity().
		Animation("run", "sprites/run.png", 4, 10).
		Animation("jump", "sprites/jump.png", 2, 6).
		Animation("death", "sprites/death.png", 4, 8))
	player.Play("run")

	hud := play.Add(engine.Text("0 m").At(60, 30))
	alive := true
	score := 0.0

	over := g.Scene("gameover")
	msg := over.Add(engine.Text("").At(400, 280))
	over.Add(engine.Text("Click to run again").At(400, 340))
	over.OnClick(func(_, _ float64) {
		alive = true
		score = 0
		player.Play("run")
		g.Restart("play")
	})

	player.OnUpdate(func(dt float64) {
		if !alive {
			return
		}
		score += 60 * dt
		hud.SetText(fmt.Sprintf("%d m", int(score)))
		if g.Key(engine.Space) && player.Grounded() {
			player.Vy = -820
			g.Sound("audios/jump.wav")
		}
		if player.Grounded() {
			player.Play("run")
		} else {
			player.Play("jump")
		}
	})

	play.Every(1.1, func() {
		if !alive {
			return
		}
		size := 26 + rand.Float64()*22
		crate := play.Add(engine.Sprite("sprites/crate.png").Size(size, size).
			At(830, 540-size/2).Tag("crate"))
		crate.Vx = -300
		crate.LifeTime(4)
	})

	player.OnCollisionWith("crate", func(_ *engine.Object) {
		if !alive {
			return
		}
		alive = false
		player.PlayOnce("death")
		g.Sound("audios/hit.wav")
		msg.SetText(fmt.Sprintf("You ran %d m", int(score)))
		play.After(1, func() { g.Go("gameover") })
	})

	g.Run("play")
}
