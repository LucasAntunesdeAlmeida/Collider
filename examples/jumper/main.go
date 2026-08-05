// Jumper: climb the platforms, grab every coin, avoid the spikes.
// Menu, animated hero, spinning coins, tiled platforms, win and lose
// screens.
//
// Run from this folder: cd examples/jumper && go run .
package main

import (
	"fmt"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

const font = "fonts/pixel.ttf"

func main() {
	g := engine.New("Jumper", 800, 600)

	// --- Menu ---
	menu := g.Scene("menu")
	menu.Music("audios/menu.wav")
	menu.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	menu.Add(engine.Text("JUMPER").At(400, 150).Font(font).TextSize(52).TextColor(engine.Yellow))
	menu.Add(engine.Text("GRAB EVERY COIN").At(400, 225).Font(font).TextSize(14))
	menu.Add(engine.Text("ARROWS MOVE   SPACE JUMPS").At(400, 500).Font(font).TextSize(12))
	menu.Add(engine.Button("PLAY").At(400, 360).Font(font).TextSize(24).Color(engine.Green)).
		OnClick(func() { g.Restart("play") })

	// --- Level ---
	play := g.Scene("play")
	play.Music("audios/level1.wav")
	play.Gravity(1800)
	play.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())

	// platform: an invisible solid slab dressed with tile sprites.
	platform := func(cx, cy, w float64, sprite string, tileH float64) {
		play.Add(engine.Rect(w, tileH, nil).At(cx, cy).Solid())
		n := int(w / 40)
		x := cx - w/2 + 20
		for range n {
			play.Add(engine.Sprite(sprite).At(x, cy).Visual())
			x += 40
		}
	}
	platform(400, 580, 800, "sprites/ground.png", 40)
	for _, p := range [][3]float64{{180, 460, 160}, {430, 350, 160}, {660, 240, 160}} {
		platform(p[0], p[1], p[2], "sprites/platform.png", 20)
	}

	coinAt := func(x, y float64) {
		c := play.Add(engine.Rect(30, 24, nil).At(x, y).Tag("coin").
			Animation("spin", "sprites/coin.png", 4, 8))
		c.Play("spin")
	}
	coinAt(180, 415)
	coinAt(430, 305)
	coinAt(660, 195)
	play.Add(engine.Sprite("sprites/spikes.png").Size(44, 20).At(560, 550).Tag("spike"))

	player := play.Add(engine.Rect(32, 46, nil).At(80, 500).WithGravity().
		Animation("walk", "sprites/player.png", 2, 8).
		Animation("idle", "sprites/idle.png", 1, 1))
	player.Play("idle")

	hud := play.Add(engine.Text("COINS 0/3").At(100, 26).Font(font).TextSize(14))
	coins := 0

	player.OnUpdate(func(dt float64) {
		moving := false
		if g.Key(engine.Left) || g.Key(engine.A) {
			player.Move(-250*dt, 0)
			moving = true
		}
		if g.Key(engine.Right) || g.Key(engine.D) {
			player.Move(250*dt, 0)
			moving = true
		}
		if moving {
			player.Play("walk")
		} else {
			player.Play("idle")
		}
		if g.Key(engine.Space) && player.Grounded() {
			player.Vy = -720
			g.Sound("audios/jump.wav")
		}
	})

	// --- End screens ---
	win := g.Scene("win")
	win.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	win.Add(engine.Text("ALL COINS!").At(400, 220).Font(font).TextSize(36).TextColor(engine.Yellow))
	win.Add(engine.Button("REPLAY").At(290, 400).Font(font).TextSize(18).Color(engine.Green)).
		OnClick(func() { g.Restart("play") })
	win.Add(engine.Button("MENU").At(540, 400).Font(font).TextSize(18)).
		OnClick(func() { g.Go("menu") })

	lose := g.Scene("lose")
	lose.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	lose.Add(engine.Text("OUCH!").At(400, 220).Font(font).TextSize(40).TextColor(engine.Orange))
	lose.Add(engine.Button("RETRY").At(290, 400).Font(font).TextSize(18).Color(engine.Green)).
		OnClick(func() { g.Restart("play") })
	lose.Add(engine.Button("MENU").At(540, 400).Font(font).TextSize(18)).
		OnClick(func() { g.Go("menu") })

	player.OnCollisionWith("coin", func(c *engine.Object) {
		c.Destroy()
		coins++
		hud.SetText(fmt.Sprintf("COINS %d/3", coins))
		g.Sound("audios/coin.wav")
		if play.Count("coin") == 0 {
			g.Sound("audios/win.wav")
			coins = 0
			hud.SetText("COINS 0/3")
			g.Go("win")
		}
	})
	player.OnCollisionWith("spike", func(_ *engine.Object) {
		g.Sound("audios/hurt.wav")
		coins = 0
		hud.SetText("COINS 0/3")
		g.Go("lose")
	})

	g.Run("menu")
}
