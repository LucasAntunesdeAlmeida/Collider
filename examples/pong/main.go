// Pong: first to five points. Two paddles, a ball that speeds up on
// every rally, a menu and a winner screen.
//
// The ball moves itself: objects have built-in velocity applied every
// frame, so constant motion costs zero lines.
//
// Run from this folder: cd examples/pong && go run .
package main

import (
	"fmt"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

const (
	font    = "fonts/pixel.ttf"
	toWin   = 5
	speedUp = 1.06
)

func main() {
	g := engine.New("Pong", 800, 600)

	// --- Menu ---
	menu := g.Scene("menu")
	menu.Music("audios/menu.wav")
	menu.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	menu.Add(engine.Text("PONG").At(400, 150).Font(font).TextSize(64).TextColor(engine.Yellow))
	menu.Add(engine.Text("FIRST TO 5").At(400, 230).Font(font).TextSize(16))
	menu.Add(engine.Text("LEFT  W / S").At(400, 470).Font(font).TextSize(12))
	menu.Add(engine.Text("RIGHT  UP / DOWN").At(400, 505).Font(font).TextSize(12))
	menu.Add(engine.Button("PLAY").At(400, 350).Font(font).TextSize(24).Color(engine.Green)).
		OnClick(func() { g.Restart("play") })

	// --- Match ---
	play := g.Scene("play")
	play.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	for y := 20; y < 600; y += 40 {
		play.Add(engine.Rect(4, 20, engine.White).At(400, float64(y)).Visual())
	}

	left := play.Add(engine.Sprite("sprites/paddle.png").At(40, 300).Solid().Tag("paddle"))
	right := play.Add(engine.Sprite("sprites/paddle.png").At(760, 300).Solid().Tag("paddle"))
	ball := play.Add(engine.Sprite("sprites/ball.png").At(400, 300))
	score := play.Add(engine.Text("0   0").At(400, 50).Font(font).TextSize(32))

	pts := [2]int{}
	serve := func(toRight bool) {
		ball.At(400, 300)
		ball.Vx, ball.Vy = 300, 220
		if !toRight {
			ball.Vx = -ball.Vx
		}
	}
	serve(true)

	// --- Winner screen ---
	over := g.Scene("over")
	over.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	winner := over.Add(engine.Text("").At(400, 220).Font(font).TextSize(36).TextColor(engine.Yellow))
	finalScore := over.Add(engine.Text("").At(400, 300).Font(font).TextSize(18))
	over.Add(engine.Button("REMATCH").At(290, 420).Font(font).TextSize(18).Color(engine.Green)).
		OnClick(func() { g.Restart("play") })
	over.Add(engine.Button("MENU").At(540, 420).Font(font).TextSize(18)).
		OnClick(func() { g.Go("menu") })

	point := func(scorer int) {
		pts[scorer]++
		score.SetText(fmt.Sprintf("%d   %d", pts[0], pts[1]))
		if pts[scorer] >= toWin {
			g.Sound("audios/win.wav")
			side := "LEFT"
			if scorer == 1 {
				side = "RIGHT"
			}
			winner.SetText(side + " WINS")
			finalScore.SetText(fmt.Sprintf("%d - %d", pts[0], pts[1]))
			pts = [2]int{}
			score.SetText("0   0")
			g.Go("over")
			return
		}
		g.Sound("audios/score.wav")
		serve(scorer == 0)
	}

	left.OnUpdate(func(dt float64) {
		if g.Key(engine.W) {
			left.Move(0, -420*dt)
		}
		if g.Key(engine.S) {
			left.Move(0, 420*dt)
		}
		left.Y = clamp(left.Y, 60, 540)
	})
	right.OnUpdate(func(dt float64) {
		if g.Key(engine.Up) {
			right.Move(0, -420*dt)
		}
		if g.Key(engine.Down) {
			right.Move(0, 420*dt)
		}
		right.Y = clamp(right.Y, 60, 540)
	})

	// Bounce off a paddle, angled by where it hit, and speed up a little.
	ball.OnCollisionWith("paddle", func(p *engine.Object) {
		ball.Vx = -ball.Vx * speedUp
		ball.Vy = (ball.Y - p.Y) * 6
		g.Sound("audios/bounce.wav")
	})

	ball.OnUpdate(func(_ float64) {
		if ball.Y < 8 || ball.Y > 592 {
			ball.Vy = -ball.Vy
			ball.Y = clamp(ball.Y, 8, 592)
			g.Sound("audios/bounce.wav")
		}
		if ball.X < 0 {
			point(1)
		}
		if ball.X > 800 {
			point(0)
		}
	})

	g.Run("menu")
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
