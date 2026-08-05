// Caves: explore a procedurally generated cave and collect every gem.
// Each round generates a brand new cave (cellular automata), places the
// player near the center and scatters gems only in reachable areas
// (BFS flood fill), so every cave is winnable.
//
// Run from this folder: cd examples/caves && go run .
package main

import (
	"fmt"
	"math/rand/v2"

	engine "github.com/LucasAntunesdeAlmeida/collider"
	"github.com/LucasAntunesdeAlmeida/collider/examples/caves/scripts"
)

const gems = 10

func main() {
	g := engine.New("Caves", 800, 600)
	play := g.Scene("play")

	player := play.Add(engine.Rect(14, 14, engine.Blue))
	hud := play.Add(engine.Text("Gems: 0").At(70, 20))

	// The whole map is rebuilt from scratch each round, so generation
	// is plain code operating on the scene, nothing engine-special.
	var mapObjects []*engine.Object
	build := func() {
		for _, o := range mapObjects {
			o.Destroy()
		}
		mapObjects = mapObjects[:0]

		grid := scripts.Generate()

		// Merge horizontal wall runs into single solid rects: same
		// geometry, far fewer objects.
		for y := range scripts.Rows {
			for x := 0; x < scripts.Cols; {
				if !grid[y][x] {
					x++
					continue
				}
				start := x
				for x < scripts.Cols && grid[y][x] {
					x++
				}
				w := float64(x-start) * scripts.Cell
				wall := play.Add(engine.Rect(w, scripts.Cell, engine.Green).
					At(float64(start)*scripts.Cell+w/2, float64(y)*scripts.Cell+scripts.Cell/2).
					Solid())
				mapObjects = append(mapObjects, wall)
			}
		}

		px, py := scripts.FindOpen(grid)
		player.At(float64(px)*scripts.Cell+scripts.Cell/2, float64(py)*scripts.Cell+scripts.Cell/2)

		open := scripts.Reachable(grid, px, py)
		for range gems {
			if len(open) == 0 {
				break
			}
			i := rand.IntN(len(open))
			c := open[i]
			open[i] = open[len(open)-1]
			open = open[:len(open)-1]
			gem := play.Add(engine.Rect(10, 10, engine.Yellow).
				At(float64(c[0])*scripts.Cell+scripts.Cell/2, float64(c[1])*scripts.Cell+scripts.Cell/2).
				Tag("gem"))
			mapObjects = append(mapObjects, gem)
		}
		hud.SetText(fmt.Sprintf("Gems: %d", play.Count("gem")))
	}
	build()

	player.OnUpdate(func(dt float64) {
		if g.Key(engine.W) || g.Key(engine.Up) {
			player.Move(0, -170*dt)
		}
		if g.Key(engine.S) || g.Key(engine.Down) {
			player.Move(0, 170*dt)
		}
		if g.Key(engine.A) || g.Key(engine.Left) {
			player.Move(-170*dt, 0)
		}
		if g.Key(engine.D) || g.Key(engine.Right) {
			player.Move(170*dt, 0)
		}
	})

	player.OnCollisionWith("gem", func(gem *engine.Object) {
		gem.Destroy()
		g.Sound("audios/gem.wav")
		hud.SetText(fmt.Sprintf("Gems: %d", play.Count("gem")))
		if play.Count("gem") == 0 {
			g.Go("win")
		}
	})

	win := g.Scene("win")
	win.Add(engine.Text("Cave cleared! Click for a new cave.").At(400, 300))
	win.OnClick(func(_, _ float64) {
		build()
		g.Go("play")
	})

	g.Run("play")
}
