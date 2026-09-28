// Horde: survive the night on a field far bigger than the window.
// WASD or arrows move the hero; chasers pour in from beyond the edges
// of the screen, wherever the hero goes.
//
// This example proves the camera: the scene follows the hero across the
// world with Camera, while the HUD and the ground color stay pinned to
// the screen with Fixed.
//
// Run from this folder: cd examples/horde && go run .
package main

import (
	"fmt"
	"math"
	"math/rand/v2"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

const (
	font  = "fonts/pixel.ttf"
	field = 2000.0 // the world spans -field..field on both axes
)

func main() {
	g := engine.New("Horde", 800, 600)

	// --- The night ---
	play := g.Scene("play")
	play.Music("audios/theme.wav")
	// The ground is pinned to the screen; the world scrolls over it.
	play.Add(engine.Rect(800, 600, engine.Black).At(400, 300).Fixed())
	scenery := []string{"sprites/tuft.png", "sprites/stone.png", "sprites/flower.png"}
	for range 500 {
		x, y := (rand.Float64()*2-1)*field, (rand.Float64()*2-1)*field
		play.Add(engine.Sprite(scenery[rand.IntN(len(scenery))]).At(x, y).Visual())
	}

	hero := play.Add(engine.Rect(36, 48, nil).At(0, 0).Tag("player").
		Animation("walk", "sprites/hero.png", 2, 8).
		Animation("idle", "sprites/idle.png", 1, 1))
	hero.Play("idle")
	clock := play.Add(engine.Text("0 S").At(400, 30).Font(font).TextSize(16).Fixed())

	survived := 0.0
	hero.OnUpdate(func(dt float64) {
		dx, dy := 0.0, 0.0
		if g.Key(engine.A) || g.Key(engine.Left) {
			dx--
		}
		if g.Key(engine.D) || g.Key(engine.Right) {
			dx++
		}
		if g.Key(engine.W) || g.Key(engine.Up) {
			dy--
		}
		if g.Key(engine.S) || g.Key(engine.Down) {
			dy++
		}
		if d := math.Hypot(dx, dy); d > 0 {
			hero.Move(dx/d*210*dt, dy/d*210*dt)
			hero.Play("walk")
		} else {
			hero.Play("idle")
		}
		hero.X = max(-field, min(field, hero.X))
		hero.Y = max(-field, min(field, hero.Y))
		play.Camera(hero.X, hero.Y)

		survived += dt
		clock.SetText(fmt.Sprintf("%.0f S", survived))
	})

	// Chasers appear on a ring just outside the view, around the hero.
	play.Every(0.6, func() {
		a := rand.Float64() * 2 * math.Pi
		c := play.Add(engine.Rect(36, 48, nil).
			At(hero.X+math.Cos(a)*560, hero.Y+math.Sin(a)*560).Tag("chaser").
			Animation("walk", "sprites/chaser.png", 2, 5))
		c.Play("walk")
		speed := 70 + rand.Float64()*60
		c.OnUpdate(func(dt float64) { c.MoveToward(hero.X, hero.Y, speed*dt) })
	})

	// --- Game over ---
	over := g.Scene("over")

	over.Add(engine.Text("THE HORDE WON").At(400, 200).Font(font).TextSize(32).TextColor(engine.Red))
	result := over.Add(engine.Text("").At(400, 280).Font(font).TextSize(16))
	over.Add(engine.Button("AGAIN").At(400, 420).Font(font).TextSize(18).Color(engine.Green)).
		OnClick(func() {
			survived = 0
			g.Restart("play")
		})

	hero.OnCollisionWith("chaser", func(*engine.Object) {
		g.Sound("audios/hurt.wav")
		result.SetText(fmt.Sprintf("SURVIVED %.0f SECONDS", survived))
		g.Go("over")
	})

	g.Run("play")
}
