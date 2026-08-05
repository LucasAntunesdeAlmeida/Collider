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

import (
	"io/fs"

	"github.com/LucasAntunesdeAlmeida/collider/internal/assets"
)

// UseAssets routes all asset loading (sprites, sounds, music, fonts)
// through a filesystem instead of the disk: pass an embed.FS and the
// game ships as a single self-contained binary.
//
//	//go:embed sprites audios
//	var content embed.FS
//
//	func main() {
//		collider.UseAssets(content)
//		...
//	}
//
// Call it before creating objects. Without it, paths load from disk,
// which is what you want during development.
func UseAssets(f fs.FS) {
	assets.SetFS(f)
}
