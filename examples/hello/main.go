// Hello, collision: the smallest complete Collider program.
// Move the blue player with the arrow keys, touch the red box for a
// collision event, click the red box to destroy it.
package main

import (
	"fmt"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

func main() {
	g := engine.New("Hello, collision", 800, 600)
	play := g.Scene("play")

	box := play.Add(engine.Rect(64, 64, engine.Red).At(400, 300))
	player := play.Add(engine.Rect(48, 48, engine.Blue).At(100, 300))

	player.OnUpdate(func(dt float64) {
		if g.Key(engine.Right) {
			player.Move(200*dt, 0)
		}
		if g.Key(engine.Left) {
			player.Move(-200*dt, 0)
		}
		if g.Key(engine.Up) {
			player.Move(0, -200*dt)
		}
		if g.Key(engine.Down) {
			player.Move(0, 200*dt)
		}
	})

	player.OnCollision(func(other *engine.Object) {
		fmt.Println("hit!")
	})

	box.OnClick(func() {
		fmt.Println("box clicked, destroying it")
		box.Destroy()
	})

	g.Run("play")
}
