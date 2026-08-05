// Arena: survive the chasers as long as you can. WASD to move, one hit
// point lost per touch, brief invulnerability after each hit.
//
// This example exists to show how to EXTEND Collider: the types in
// scripts/ embed *engine.Object and add their own fields and methods
// (HP, TakeDamage, Knockback, chase speed). The engine sees normal
// objects; your game code sees Players and Chasers.
//
// Run from this folder: cd examples/arena && go run .
package main

import (
	"fmt"

	engine "github.com/LucasAntunesdeAlmeida/collider"
	"github.com/LucasAntunesdeAlmeida/collider/examples/arena/scripts"
)

func main() {
	g := engine.New("Arena", 800, 600)
	play := g.Scene("play")

	player := scripts.NewPlayer(g, play)
	hud := play.Add(engine.Text("HP: 3").At(60, 30))

	play.Every(2, func() { scripts.NewChaser(play, player) })

	player.OnCollisionWith("chaser", func(c *engine.Object) {
		player.TakeDamage(1)
		player.Knockback(c.X, c.Y)
		hud.SetText(fmt.Sprintf("HP: %d", player.HP))
		if player.HP <= 0 {
			g.Go("gameover")
		}
	})

	over := g.Scene("gameover")
	over.Add(engine.Text("They got you. Click to try again.").At(400, 300))
	over.OnClick(func(_, _ float64) {
		player.Reset()
		hud.SetText("HP: 3")
		g.Restart("play")
	})

	g.Run("play")
}
