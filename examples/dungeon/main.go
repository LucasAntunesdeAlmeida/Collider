// Dungeon: four connected rooms, a key, a locked door, a treasure
// chest. Walk off a screen edge to enter the neighboring room. Menu,
// animated hero, drawn props, tiled walls.
//
// Run from this folder: cd examples/dungeon && go run .
package main

import (
	"fmt"

	engine "github.com/LucasAntunesdeAlmeida/collider"
	"github.com/LucasAntunesdeAlmeida/collider/examples/dungeon/scripts"
)

const font = "fonts/pixel.ttf"

func main() {
	g := engine.New("Dungeon", 800, 600)

	// --- Menu ---
	menu := g.Scene("menu")
	menu.Music("audios/dungeon.wav")
	menu.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	menu.Add(engine.Text("DUNGEON").At(400, 150).Font(font).TextSize(48).TextColor(engine.Orange))
	menu.Add(engine.Text("FIND THE KEY  OPEN THE VAULT").At(400, 225).Font(font).TextSize(12))
	menu.Add(engine.Text("WASD OR ARROWS   EDGES LEAD ONWARD").At(400, 500).Font(font).TextSize(10))
	menu.Add(engine.Button("ENTER").At(400, 360).Font(font).TextSize(22).Color(engine.Orange)).
		OnClick(func() { g.Go("play") })

	// --- Dungeon ---
	play := g.Scene("play")
	play.Music("audios/dungeon.wav")
	play.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())

	player := play.Add(engine.Rect(28, 40, nil).At(120, 120).Tag("player").
		Animation("walk", "sprites/player.png", 2, 8))
	player.Play("walk")
	hud := play.Add(engine.Text("").At(140, 20).Font(font).TextSize(12))

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
		cell := scripts.Cell
		at := func(x, y int) (float64, float64) {
			return float64(x)*cell + cell/2, float64(y)*cell + cell/2
		}
		add := func(o *engine.Object) {
			mapObjs = append(mapObjs, play.Add(o))
		}
		for y, row := range layout {
			for x := 0; x < len(row); {
				switch row[x] {
				case '#':
					start := x
					for x < len(row) && row[x] == '#' {
						x++
					}
					w := float64(x-start) * cell
					cx, cy := at(start, y)
					add(engine.Rect(w, cell, nil).At(cx+w/2-cell/2, cy).Solid())
					for i := start; i < x; i++ {
						tx, ty := at(i, y)
						add(engine.Sprite("sprites/wall.png").At(tx, ty).Visual())
					}
					continue
				case 'K':
					if !hasKey {
						cx, cy := at(x, y)
						add(engine.Sprite("sprites/key.png").At(cx, cy).Tag("key"))
					}
				case 'D':
					if !doorOpen {
						cx, cy := at(x, y)
						add(engine.Sprite("sprites/door.png").Size(cell, cell).
							At(cx, cy).Solid().Tag("door"))
					}
				case 'T':
					cx, cy := at(x, y)
					add(engine.Sprite("sprites/chest.png").At(cx, cy).Tag("treasure"))
				}
				x++
			}
		}
		key := "NO"
		if hasKey {
			key = "YES"
		}
		hud.SetText(fmt.Sprintf("ROOM %d.%d  KEY %s", cur[0], cur[1], key))
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

	// --- Treasure ---
	win := g.Scene("win")
	win.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	win.Add(engine.Text("TREASURE!").At(400, 200).Font(font).TextSize(40).TextColor(engine.Yellow))
	win.Add(engine.Text("THE VAULT IS YOURS").At(400, 280).Font(font).TextSize(14))
	win.Add(engine.Button("EXPLORE AGAIN").At(400, 420).Font(font).TextSize(16).Color(engine.Orange)).
		OnClick(func() {
			hasKey = false
			doorOpen = false
			cur = [2]int{0, 0}
			player.At(120, 120)
			build()
			g.Go("play")
		})

	player.OnCollisionWith("treasure", func(_ *engine.Object) {
		g.Sound("audios/win.wav")
		g.Go("win")
	})

	g.Run("menu")
}
