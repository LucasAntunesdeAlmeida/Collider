// Package collider is a code-first 2D game engine focused on collisions
// and events.
//
// You import the library, describe your objects, attach events, and you
// have a game: no editor, no project files, no boilerplate game loop.
//
//	g := collider.New("My Game", 800, 600)
//	play := g.Scene("play")
//
//	player := play.Add(collider.Rect(48, 48, collider.Blue).At(100, 300))
//	player.OnCollisionWith("enemy", func(e *collider.Object) {
//		g.Go("gameover")
//	})
//
//	g.Run("play")
//
// See README.md for the full spec and five complete example games.
package collider
