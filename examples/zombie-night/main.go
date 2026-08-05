// Zombie Night: the stress test for runtime spawning. Objects are created
// and destroyed constantly, so collision rules are declared once at the
// scene level, by tag pair, and apply to every current and future object.
//
// Run from this folder so the assets resolve: cd examples/zombie-night && go run .
package main

import (
	"fmt"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

func main() {
	g := engine.New("Zombie Night", 800, 600)
	play := g.Scene("play")
	play.Music("audios/dread.wav")

	player := play.Add(engine.Sprite("sprites/hero.png").At(400, 300))
	kills := 0
	hud := play.Add(engine.Text("Kills: 0").At(70, 30))

	player.OnUpdate(func(dt float64) {
		if g.Key(engine.W) {
			player.Move(0, -220*dt)
		}
		if g.Key(engine.S) {
			player.Move(0, 220*dt)
		}
		if g.Key(engine.A) {
			player.Move(-220*dt, 0)
		}
		if g.Key(engine.D) {
			player.Move(220*dt, 0)
		}
	})

	// A zombie every 1.5 seconds, from a random screen edge, walking at the player.
	play.Every(1.5, func() {
		z := play.Add(engine.Sprite("sprites/zombie.png").AtEdge().Tag("zombie"))
		z.OnUpdate(func(dt float64) { z.MoveToward(player.X, player.Y, 60*dt) })
	})

	// Click anywhere to shoot toward the cursor.
	play.OnClick(func(x, y float64) {
		b := play.Add(engine.Rect(6, 6, engine.Yellow).At(player.X, player.Y).Tag("bullet"))
		b.VelocityToward(x, y, 600)
		b.LifeTime(2)
		g.Sound("audios/shot.wav")
	})

	// Scene-level rules: apply to every zombie and bullet, present or future.
	play.OnCollision("zombie", "bullet", func(z, b *engine.Object) {
		z.Destroy()
		b.Destroy()
		kills++
		hud.SetText(fmt.Sprintf("Kills: %d", kills))
	})
	player.OnCollisionWith("zombie", func(_ *engine.Object) { g.Go("gameover") })

	over := g.Scene("gameover")
	over.Add(engine.Text("They got you.").At(400, 280))
	again := over.Add(engine.Sprite("sprites/again.png").At(400, 360))
	again.OnClick(func() {
		kills = 0
		g.Restart("play")
	})

	g.Run("play")
}
