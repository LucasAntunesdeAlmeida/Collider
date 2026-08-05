// Pong: two paddles, a ball, a score. The ball moves itself, because
// objects have built-in velocity applied every frame.
//
// Run from this folder so the assets resolve: cd examples/pong && go run .
package main

import (
	"fmt"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

func main() {
	g := engine.New("Pong", 800, 600)
	play := g.Scene("play")

	left := play.Add(engine.Rect(20, 100, engine.White).At(40, 300).Solid())
	right := play.Add(engine.Rect(20, 100, engine.White).At(760, 300).Solid())
	ball := play.Add(engine.Rect(16, 16, engine.White).At(400, 300))
	score := play.Add(engine.Text("0 : 0").At(400, 40))

	pts := [2]int{}
	serve := func() { ball.At(400, 300); ball.Vx, ball.Vy = 300, 200 }
	serve()

	left.OnUpdate(func(dt float64) {
		if g.Key(engine.W) {
			left.Move(0, -400*dt)
		}
		if g.Key(engine.S) {
			left.Move(0, 400*dt)
		}
	})
	right.OnUpdate(func(dt float64) {
		if g.Key(engine.Up) {
			right.Move(0, -400*dt)
		}
		if g.Key(engine.Down) {
			right.Move(0, 400*dt)
		}
	})

	ball.OnCollision(func(_ *engine.Object) {
		ball.Vx = -ball.Vx
		g.Sound("bounce.wav")
	})

	ball.OnUpdate(func(_ float64) {
		if ball.Y < 0 || ball.Y > 600 {
			ball.Vy = -ball.Vy
		}
		if ball.X < 0 {
			pts[1]++
			serve()
		}
		if ball.X > 800 {
			pts[0]++
			serve()
		}
		score.SetText(fmt.Sprintf("%d : %d", pts[0], pts[1]))
	})

	g.Run("play")
}
