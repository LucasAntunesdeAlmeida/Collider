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

const (
	font = "fonts/pixel.ttf"
	gems = 10
)

func main() {
	g := engine.New("Caves", 800, 600)

	// --- Menu ---
	menu := g.Scene("menu")
	menu.Music("audios/theme.wav")
	menu.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	menu.Add(engine.Text("CAVES").At(400, 150).Font(font).TextSize(52).TextColor(engine.Orange))
	menu.Add(engine.Text("EVERY CAVE IS BRAND NEW").At(400, 225).Font(font).TextSize(12))
	menu.Add(engine.Text("WASD OR ARROWS TO EXPLORE").At(400, 500).Font(font).TextSize(12))
	menu.Add(engine.Button("DESCEND").At(400, 360).Font(font).TextSize(22).Color(engine.Orange)).
		OnClick(func() { g.Go("play") })

	// --- Cave ---
	play := g.Scene("play")
	play.Music("audios/theme.wav")
	play.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())

	player := play.Add(engine.Rect(26, 38, nil).Tag("player").
		Animation("walk", "sprites/player.png", 2, 8))
	player.Play("walk")
	hud := play.Add(engine.Text("").At(100, 20).Font(font).TextSize(12))

	// The whole map is rebuilt from scratch each round: generation is
	// plain code operating on the scene, nothing engine-special.
	var mapObjects []*engine.Object
	build := func() {
		for _, o := range mapObjects {
			o.Destroy()
		}
		mapObjects = mapObjects[:0]

		grid := scripts.Generate()
		cell := scripts.Cell
		at := func(x, y int) (float64, float64) {
			return float64(x)*cell + cell/2, float64(y)*cell + cell/2
		}

		// Invisible merged solids for physics, rock tiles for looks.
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
				w := float64(x-start) * cell
				cx, cy := at(start, y)
				wall := play.Add(engine.Rect(w, cell, nil).
					At(cx+w/2-cell/2, cy).Solid())
				mapObjects = append(mapObjects, wall)
				for i := start; i < x; i++ {
					tx, ty := at(i, y)
					mapObjects = append(mapObjects, play.Add(
						engine.Sprite("sprites/rock.png").At(tx, ty).Visual()))
				}
			}
		}

		px, py := scripts.FindOpen(grid)
		sx, sy := at(px, py)
		player.At(sx, sy)

		open := scripts.Reachable(grid, px, py)
		for range gems {
			if len(open) == 0 {
				break
			}
			i := rand.IntN(len(open))
			c := open[i]
			open[i] = open[len(open)-1]
			open = open[:len(open)-1]
			gx, gy := at(c[0], c[1])
			gem := play.Add(engine.Rect(20, 16, nil).At(gx, gy).Tag("gem").
				Animation("sparkle", "sprites/gem.png", 2, 4))
			gem.Play("sparkle")
			mapObjects = append(mapObjects, gem)
		}
		hud.SetText(fmt.Sprintf("GEMS %d", play.Count("gem")))
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
		hud.SetText(fmt.Sprintf("GEMS %d", play.Count("gem")))
		if play.Count("gem") == 0 {
			g.Sound("audios/win.wav")
			g.Go("win")
		}
	})

	// --- Cleared ---
	win := g.Scene("win")
	win.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	win.Add(engine.Text("CAVE CLEARED").At(400, 220).Font(font).TextSize(32).TextColor(engine.Yellow))
	win.Add(engine.Button("NEW CAVE").At(290, 400).Font(font).TextSize(18).Color(engine.Orange)).
		OnClick(func() {
			build()
			g.Go("play")
		})
	win.Add(engine.Button("MENU").At(540, 400).Font(font).TextSize(18)).
		OnClick(func() { g.Go("menu") })

	g.Run("menu")
}
