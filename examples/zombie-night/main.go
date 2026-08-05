// Zombie Night: survive the horde. WASD moves the hero, click shoots
// toward the cursor, zombies shamble in from the edges faster and
// faster. Menu, animated characters, kill counter, game over screen.
//
// The stress test for runtime spawning: collision rules are declared
// once at the scene level, by tag pair, and apply to every current and
// future object.
//
// Run from this folder: cd examples/zombie-night && go run .
package main

import (
	"fmt"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

const font = "fonts/pixel.ttf"

func main() {
	g := engine.New("Zombie Night", 800, 600)

	// --- Menu ---
	menu := g.Scene("menu")
	menu.Music("audios/menu.wav")
	menu.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	menu.Add(engine.Text("ZOMBIE NIGHT").At(400, 150).Font(font).TextSize(40).TextColor(engine.Green))
	menu.Add(engine.Text("SURVIVE").At(400, 225).Font(font).TextSize(14))
	menu.Add(engine.Text("WASD MOVES   CLICK SHOOTS").At(400, 500).Font(font).TextSize(12))
	menu.Add(engine.Button("PLAY").At(400, 360).Font(font).TextSize(24).Color(engine.Green)).
		OnClick(func() { g.Restart("play") })

	// --- Night ---
	play := g.Scene("play")
	play.Music("audios/dread.wav")
	play.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())

	player := play.Add(engine.Rect(32, 45, nil).At(400, 300).Tag("player").
		Animation("walk", "sprites/hero.png", 2, 8).
		Animation("idle", "sprites/hero-idle.png", 1, 1))
	player.Play("idle")

	hud := play.Add(engine.Text("KILLS 0").At(90, 26).Font(font).TextSize(16))
	kills := 0

	player.OnUpdate(func(dt float64) {
		moving := false
		if g.Key(engine.W) {
			player.Move(0, -220*dt)
			moving = true
		}
		if g.Key(engine.S) {
			player.Move(0, 220*dt)
			moving = true
		}
		if g.Key(engine.A) {
			player.Move(-220*dt, 0)
			moving = true
		}
		if g.Key(engine.D) {
			player.Move(220*dt, 0)
			moving = true
		}
		if moving {
			player.Play("walk")
		} else {
			player.Play("idle")
		}
	})

	// The horde: spawn rate is fixed, but each zombie is a bit faster
	// than the last night's average as kills climb.
	play.Every(1.4, func() {
		z := play.Add(engine.Rect(32, 45, nil).AtEdge().Tag("zombie").
			Animation("walk", "sprites/zombie.png", 2, 5))
		z.Play("walk")
		speed := 55 + float64(kills)*1.5
		z.OnUpdate(func(dt float64) { z.MoveToward(player.X, player.Y, speed*dt) })
	})

	play.OnClick(func(x, y float64) {
		b := play.Add(engine.Sprite("sprites/bullet.png").At(player.X, player.Y).Tag("bullet"))
		b.VelocityToward(x, y, 620)
		b.LifeTime(2)
		g.Sound("audios/shot.wav")
	})

	play.OnCollision("zombie", "bullet", func(z, b *engine.Object) {
		z.Destroy()
		b.Destroy()
		kills++
		hud.SetText(fmt.Sprintf("KILLS %d", kills))
		g.Sound("audios/kill.wav")
	})

	// --- Game over ---
	over := g.Scene("over")
	over.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	over.Add(engine.Text("THEY GOT YOU").At(400, 200).Font(font).TextSize(32).TextColor(engine.Red))
	result := over.Add(engine.Text("").At(400, 280).Font(font).TextSize(16))
	over.Add(engine.Button("AGAIN").At(290, 420).Font(font).TextSize(18).Color(engine.Green)).
		OnClick(func() {
			kills = 0
			hud.SetText("KILLS 0")
			g.Restart("play")
		})
	over.Add(engine.Button("MENU").At(540, 420).Font(font).TextSize(18)).
		OnClick(func() {
			kills = 0
			hud.SetText("KILLS 0")
			g.Go("menu")
		})

	player.OnCollisionWith("zombie", func(_ *engine.Object) {
		g.Sound("audios/death.wav")
		result.SetText(fmt.Sprintf("%d KILLS", kills))
		g.Go("over")
	})

	g.Run("menu")
}
