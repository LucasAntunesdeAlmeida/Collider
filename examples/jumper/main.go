// Jumper: the genre that demands collision response, not just detection.
// Solid objects push dynamic objects out, Grounded reports when the
// player can jump, and gravity is one line on the scene.
//
// Run from this folder so the assets resolve: cd examples/jumper && go run .
package main

import engine "github.com/LucasAntunesdeAlmeida/collider"

func main() {
	g := engine.New("Jumper", 800, 600)
	play := g.Scene("play")
	play.Music("level1.wav")
	play.Gravity(1800)

	// Level geometry: solid means the engine resolves the collision for you.
	play.Add(engine.Rect(800, 40, engine.Green).At(400, 580).Solid())
	for _, p := range [][2]float64{{200, 450}, {430, 340}, {650, 230}} {
		play.Add(engine.Rect(140, 20, engine.Green).At(p[0], p[1]).Solid())
	}
	for _, p := range [][2]float64{{200, 410}, {430, 300}, {650, 190}} {
		play.Add(engine.Sprite("coin.png").At(p[0], p[1]).Tag("coin"))
	}
	play.Add(engine.Sprite("spikes.png").At(550, 555).Tag("spike"))

	player := play.Add(engine.Sprite("player.png").At(80, 500).WithGravity())

	player.OnUpdate(func(dt float64) {
		if g.Key(engine.Left) {
			player.Move(-250*dt, 0)
		}
		if g.Key(engine.Right) {
			player.Move(250*dt, 0)
		}
		if g.Key(engine.Space) && player.Grounded() {
			player.Vy = -700
		}
	})

	player.OnCollisionWith("coin", func(c *engine.Object) {
		c.Destroy()
		g.Sound("coin.wav")
		if play.Count("coin") == 0 {
			g.Go("win")
		}
	})
	player.OnCollisionWith("spike", func(_ *engine.Object) { g.Go("gameover") })

	win := g.Scene("win")
	win.Add(engine.Text("You win!").At(400, 250))
	replay := win.Add(engine.Sprite("replay.png").At(400, 350))
	replay.OnClick(func() { g.Restart("play") })

	over := g.Scene("gameover")
	over.Add(engine.Text("Ouch.").At(400, 250))
	retry := over.Add(engine.Sprite("retry.png").At(400, 350))
	retry.OnClick(func() { g.Restart("play") })

	g.Run("play")
}
