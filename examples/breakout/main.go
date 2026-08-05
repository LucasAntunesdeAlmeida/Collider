// Breakout: clear fifty bricks with three lives. Menu, lives, a ball
// that angles off the paddle, and win and lose screens.
//
// Run from this folder: cd examples/breakout && go run .
package main

import (
	"fmt"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

const font = "fonts/pixel.ttf"

var brickColors = []string{"red", "orange", "yellow", "green", "blue"}

func main() {
	g := engine.New("Breakout", 800, 600)

	// --- Menu ---
	menu := g.Scene("menu")
	menu.Music("audios/menu.wav")
	menu.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	menu.Add(engine.Text("BREAKOUT").At(400, 150).Font(font).TextSize(48).TextColor(engine.Yellow))
	menu.Add(engine.Text("CLEAR EVERY BRICK").At(400, 230).Font(font).TextSize(14))
	menu.Add(engine.Text("MOUSE MOVES THE PADDLE").At(400, 490).Font(font).TextSize(12))
	menu.Add(engine.Button("PLAY").At(400, 350).Font(font).TextSize(24).Color(engine.Green)).
		OnClick(func() { g.Restart("play") })

	// --- Game ---
	play := g.Scene("play")
	play.Music("audios/action.wav")
	play.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())

	for row := range 5 {
		for col := range 10 {
			play.Add(engine.Sprite("sprites/brick-"+brickColors[row]+".png").
				At(60+float64(col)*76, 90+float64(row)*34).Tag("brick"))
		}
	}

	paddle := play.Add(engine.Sprite("sprites/paddle.png").At(400, 560).Solid().Tag("paddle"))
	ball := play.Add(engine.Sprite("sprites/ball.png").At(400, 430))
	hud := play.Add(engine.Text("").At(120, 30).Font(font).TextSize(16))

	lives := 3
	score := 0
	updateHUD := func() {
		hud.SetText(fmt.Sprintf("SCORE %d   LIVES %d", score, lives))
	}
	launch := func() {
		ball.At(paddle.X, 430)
		ball.Vx, ball.Vy = 240, -340
	}
	launch()
	updateHUD()

	// --- End screens ---
	over := g.Scene("over")
	over.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	title := over.Add(engine.Text("").At(400, 210).Font(font).TextSize(40).TextColor(engine.Yellow))
	result := over.Add(engine.Text("").At(400, 290).Font(font).TextSize(16))
	over.Add(engine.Button("AGAIN").At(290, 420).Font(font).TextSize(18).Color(engine.Green)).
		OnClick(func() { g.Restart("play") })
	over.Add(engine.Button("MENU").At(540, 420).Font(font).TextSize(18)).
		OnClick(func() { g.Go("menu") })

	finish := func(won bool) {
		if won {
			g.Sound("audios/win.wav")
			title.SetText("CLEARED")
		} else {
			g.Sound("audios/lose.wav")
			title.SetText("GAME OVER")
		}
		result.SetText(fmt.Sprintf("SCORE %d", score))
		lives, score = 3, 0
		updateHUD()
		g.Go("over")
	}

	paddle.OnUpdate(func(_ float64) {
		x, _ := g.Mouse()
		paddle.X = clamp(x, 60, 740)
	})

	ball.OnCollisionWith("brick", func(brick *engine.Object) {
		brick.Destroy()
		ball.Vy = -ball.Vy
		score += 10
		updateHUD()
		g.Sound("audios/break.wav")
		if play.Count("brick") == 0 {
			finish(true)
		}
	})

	// Angle the bounce by where the ball hit the paddle.
	ball.OnCollisionWith("paddle", func(p *engine.Object) {
		ball.Vy = -absF(ball.Vy)
		ball.Vx = (ball.X - p.X) * 7
		g.Sound("audios/bounce.wav")
	})

	ball.OnUpdate(func(_ float64) {
		if ball.X < 8 || ball.X > 792 {
			ball.Vx = -ball.Vx
			ball.X = clamp(ball.X, 8, 792)
		}
		if ball.Y < 8 {
			ball.Vy = -ball.Vy
			ball.Y = 8
		}
		if ball.Y > 600 {
			lives--
			updateHUD()
			if lives <= 0 {
				finish(false)
				return
			}
			g.Sound("audios/lose.wav")
			launch()
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

func absF(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
