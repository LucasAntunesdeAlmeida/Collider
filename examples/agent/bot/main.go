// A bot that plays Gem Rush headless, using only what an external
// agent would have: observations (tags, positions) in, key presses
// out. No window, no reading game internals, deterministic.
//
// Run: cd examples/agent/bot && go run .
package main

import (
	"fmt"
	"math"
	"os"

	engine "github.com/LucasAntunesdeAlmeida/collider"
	"github.com/LucasAntunesdeAlmeida/collider/examples/agent/game"
)

func main() {
	os.Chdir("..") // assets resolve relative to the game's folder

	g := game.New()
	g.Headless("play")

	steps := 0
	obs := g.Observe()
	for obs.Scene == "play" && steps < 60*60 {
		var player *engine.ObjectObs
		var gems []engine.ObjectObs
		for i, o := range obs.Objects {
			switch o.Tag {
			case "player":
				player = &obs.Objects[i]
			case "gem":
				gems = append(gems, o)
			}
		}
		if player == nil || len(gems) == 0 {
			obs = g.Step(engine.Action{})
			steps++
			continue
		}

		// Chase the nearest gem.
		target := gems[0]
		best := math.Hypot(target.X-player.X, target.Y-player.Y)
		for _, gem := range gems[1:] {
			if d := math.Hypot(gem.X-player.X, gem.Y-player.Y); d < best {
				best, target = d, gem
			}
		}

		var keys []engine.Key
		if target.X < player.X-4 {
			keys = append(keys, engine.A)
		}
		if target.X > player.X+4 {
			keys = append(keys, engine.D)
		}
		if target.Y < player.Y-4 {
			keys = append(keys, engine.W)
		}
		if target.Y > player.Y+4 {
			keys = append(keys, engine.S)
		}

		obs = g.Step(engine.Action{Keys: keys})
		steps++
	}

	fmt.Printf("finished in scene %q after %d steps (%.1f game seconds)\n",
		obs.Scene, steps, float64(steps)/60)
}
