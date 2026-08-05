// Arena: survive the chasers. WASD or arrows, three hearts, brief
// invulnerability after each hit, knockback on contact. Menu, animated
// characters, heart HUD, game over screen.
//
// This example shows how to EXTEND Collider: the types in scripts/
// embed *engine.Object and add their own fields and methods.
//
// Run from this folder: cd examples/arena && go run .
package main

import (
	"fmt"

	engine "github.com/LucasAntunesdeAlmeida/collider"
	"github.com/LucasAntunesdeAlmeida/collider/examples/arena/scripts"
)

const font = "fonts/pixel.ttf"

func main() {
	g := engine.New("Arena", 800, 600)

	// --- Menu ---
	menu := g.Scene("menu")
	menu.Music("audios/theme.wav")
	menu.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	menu.Add(engine.Text("ARENA").At(400, 150).Font(font).TextSize(52).TextColor(engine.Red))
	menu.Add(engine.Text("SURVIVE THE CHASERS").At(400, 225).Font(font).TextSize(14))
	menu.Add(engine.Text("WASD OR ARROWS TO MOVE").At(400, 500).Font(font).TextSize(12))
	menu.Add(engine.Button("FIGHT").At(400, 360).Font(font).TextSize(24).Color(engine.Red)).
		OnClick(func() { g.Restart("play") })

	// --- Arena ---
	play := g.Scene("play")
	play.Music("audios/theme.wav")
	play.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())

	player := scripts.NewPlayer(g, play)

	hearts := make([]*engine.Object, 3)
	for i := range hearts {
		hearts[i] = play.Add(engine.Sprite("sprites/heart.png").
			At(40+float64(i)*36, 30).Visual())
	}
	timeHud := play.Add(engine.Text("0 S").At(740, 26).Font(font).TextSize(14))

	survived := 0.0
	alive := true
	timeHud.OnUpdate(func(dt float64) {
		if alive {
			survived += dt
			timeHud.SetText(fmt.Sprintf("%.0f S", survived))
		}
	})

	play.Every(2, func() {
		if alive {
			scripts.NewChaser(play, player)
		}
	})

	// --- Game over ---
	over := g.Scene("over")
	over.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	over.Add(engine.Text("THEY GOT YOU").At(400, 200).Font(font).TextSize(32).TextColor(engine.Red))
	result := over.Add(engine.Text("").At(400, 280).Font(font).TextSize(16))
	reset := func() {
		player.Reset()
		survived = 0
		alive = true
	}
	over.Add(engine.Button("AGAIN").At(290, 420).Font(font).TextSize(18).Color(engine.Red)).
		OnClick(func() {
			reset()
			g.Restart("play")
		})
	over.Add(engine.Button("MENU").At(540, 420).Font(font).TextSize(18)).
		OnClick(func() {
			reset()
			g.Go("menu")
		})

	player.OnCollisionWith("chaser", func(c *engine.Object) {
		if !alive || !player.TakeDamage(1) {
			return
		}
		player.Knockback(c.X, c.Y)
		if player.HP >= 0 && player.HP < len(hearts) {
			hearts[player.HP].Destroy() // Restart restores setup objects
		}
		g.Sound("audios/hurt.wav")
		if player.HP <= 0 {
			alive = false
			g.Sound("audios/death.wav")
			result.SetText(fmt.Sprintf("SURVIVED %.0f SECONDS", survived))
			play.After(0.6, func() { g.Go("over") })
		}
	})

	g.Run("menu")
}
