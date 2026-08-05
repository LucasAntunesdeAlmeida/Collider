// Dungeon: four connected rooms, a key, a locked door, a treasure.
// Walk off a screen edge to enter the neighboring room.
//
// This example proves multiple screens: rooms are ASCII layouts in
// scripts/, one scene rebuilds its content on each transition, and
// world state (key collected, door opened) is plain Go variables that
// survive room changes.
//
// Run from this folder: cd examples/dungeon && go run .
package main

import (
	"fmt"

	engine "github.com/LucasAntunesdeAlmeida/collider"
	"github.com/LucasAntunesdeAlmeida/collider/examples/dungeon/scripts"
)

func main() {
	g := engine.New("Dungeon", 800, 600)
	play := g.Scene("play")
	play.Music("audios/dungeon.wav")

	player := play.Add(engine.Rect(22, 22, engine.Blue).At(120, 120))
	hud := play.Add(engine.Text("").At(110, 22).TextSize(18))

	cur := [2]int{0, 0}
	hasKey := false
	doorOpen := false

	var mapObjs []*engine.Object
	build := func() {
		for _, o := range mapObjs {
			o.Destroy()
		}
		mapObjs = mapObjs[:0]

		layout := scripts.Rooms[cur]
		at := func(x, y int) (float64, float64) {
			return float64(x)*scripts.Cell + scripts.Cell/2, float64(y)*scripts.Cell + scripts.Cell/2
		}
		for y, row := range layout {
			for x := 0; x < len(row); {
				switch row[x] {
				case '#':
					start := x
					for x < len(row) && row[x] == '#' {
						x++
					}
					w := float64(x-start) * scripts.Cell
					cx, cy := at(start, y)
					wall := play.Add(engine.Rect(w, scripts.Cell, engine.Green).
						At(cx+w/2-scripts.Cell/2, cy).Solid())
					mapObjs = append(mapObjs, wall)
					continue
				case 'K':
					if !hasKey {
						cx, cy := at(x, y)
						mapObjs = append(mapObjs, play.Add(
							engine.Rect(16, 16, engine.Yellow).At(cx, cy).Tag("key")))
					}
				case 'D':
					if !doorOpen {
						cx, cy := at(x, y)
						mapObjs = append(mapObjs, play.Add(
							engine.Rect(scripts.Cell, scripts.Cell, engine.Orange).
								At(cx, cy).Solid().Tag("door")))
					}
				case 'T':
					cx, cy := at(x, y)
					mapObjs = append(mapObjs, play.Add(
						engine.Rect(24, 24, engine.Red).At(cx, cy).Tag("treasure")))
				}
				x++
			}
		}
		key := "no"
		if hasKey {
			key = "yes"
		}
		hud.SetText(fmt.Sprintf("Room %d,%d   Key: %s", cur[0], cur[1], key))
	}
	build()

	enter := func(next [2]int, px, py float64) {
		if _, ok := scripts.Rooms[next]; !ok {
			return
		}
		cur = next
		player.At(px, py)
		build()
	}

	player.OnUpdate(func(dt float64) {
		if g.Key(engine.W) || g.Key(engine.Up) {
			player.Move(0, -220*dt)
		}
		if g.Key(engine.S) || g.Key(engine.Down) {
			player.Move(0, 220*dt)
		}
		if g.Key(engine.A) || g.Key(engine.Left) {
			player.Move(-220*dt, 0)
		}
		if g.Key(engine.D) || g.Key(engine.Right) {
			player.Move(220*dt, 0)
		}

		switch {
		case player.X < 0:
			enter([2]int{cur[0] - 1, cur[1]}, 788, player.Y)
		case player.X > 800:
			enter([2]int{cur[0] + 1, cur[1]}, 12, player.Y)
		case player.Y < 0:
			enter([2]int{cur[0], cur[1] - 1}, player.X, 588)
		case player.Y > 600:
			enter([2]int{cur[0], cur[1] + 1}, player.X, 12)
		}
	})

	player.OnCollisionWith("key", func(k *engine.Object) {
		k.Destroy()
		hasKey = true
		g.Sound("audios/key.wav")
		build()
	})
	player.OnCollisionWith("door", func(d *engine.Object) {
		if hasKey {
			d.Destroy()
			doorOpen = true
			g.Sound("audios/door.wav")
		}
	})
	player.OnCollisionWith("treasure", func(_ *engine.Object) {
		g.Sound("audios/win.wav")
		g.Go("win")
	})

	win := g.Scene("win")
	win.Add(engine.Text("You found the treasure!").At(400, 270).TextSize(32))
	win.Add(engine.Text("Click to explore again.").At(400, 330))
	win.OnClick(func(_, _ float64) {
		hasKey = false
		doorOpen = false
		cur = [2]int{0, 0}
		player.At(120, 120)
		build()
		g.Go("play")
	})

	g.Run("play")
}
