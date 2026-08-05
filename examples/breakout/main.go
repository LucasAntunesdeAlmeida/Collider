// Breakout: grid spawning, a win condition, mouse-follow control, and the
// full scene flow from menu to play to win or lose, all using the same
// objects and events model.
//
// Run from this folder so the assets resolve: cd examples/breakout && go run .
package main

import engine "github.com/LucasAntunesdeAlmeida/collider"

func main() {
	g := engine.New("Breakout", 800, 600)

	menu := g.Scene("menu")
	menu.Music("audios/menu.wav")
	menu.Add(engine.Text("BREAKOUT").At(400, 200))
	start := menu.Add(engine.Sprite("sprites/start.png").At(400, 350))
	start.OnClick(func() { g.Go("play") })

	play := g.Scene("play")
	play.Music("audios/action.wav")

	colors := []engine.Color{engine.Red, engine.Orange, engine.Yellow, engine.Green, engine.Blue}
	for row := range 5 {
		for col := range 10 {
			play.Add(engine.Rect(70, 24, colors[row]).
				At(60+float64(col)*76, 60+float64(row)*30).Tag("brick"))
		}
	}

	paddle := play.Add(engine.Rect(120, 20, engine.White).At(400, 560).Solid().Tag("paddle"))
	ball := play.Add(engine.Rect(14, 14, engine.White).At(400, 400))
	ball.Vx, ball.Vy = 250, -350

	paddle.OnUpdate(func(_ float64) {
		x, _ := g.Mouse()
		paddle.X = x
	})

	ball.OnCollisionWith("brick", func(brick *engine.Object) {
		brick.Destroy()
		ball.Vy = -ball.Vy
		g.Sound("audios/break.wav")
		if play.Count("brick") == 0 {
			g.Go("win")
		}
	})
	ball.OnCollisionWith("paddle", func(_ *engine.Object) { ball.Vy = -ball.Vy })

	ball.OnUpdate(func(_ float64) {
		if ball.X < 0 || ball.X > 800 {
			ball.Vx = -ball.Vx
		}
		if ball.Y < 0 {
			ball.Vy = -ball.Vy
		}
		if ball.Y > 600 {
			g.Go("gameover")
		}
	})

	win := g.Scene("win")
	win.Add(engine.Text("Cleared!").At(400, 300))
	win.OnClick(func(_, _ float64) { g.Go("menu") })

	over := g.Scene("gameover")
	over.Add(engine.Text("Game Over").At(400, 300))
	over.OnClick(func(_, _ float64) { g.Restart("play") })

	g.Run("menu")
}
