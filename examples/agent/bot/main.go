// The headless run: the same Decide pilot plays Gem Rush with no window
// at all, via Headless and Step. This is the CI shape of agent play:
// drive the game, assert it reaches "win", print how long it took.
//
// Run: cd examples/agent/bot && go run .
package main

import (
	"fmt"
	"os"

	"github.com/LucasAntunesdeAlmeida/collider/examples/agent/game"
)

func main() {
	os.Chdir("..") // assets resolve relative to the game's folder

	g := game.New()
	g.Headless("play")

	steps := 0
	obs := g.Observe()
	for obs.Scene == "play" && steps < 60*60 {
		obs = g.Step(game.Decide(obs))
		steps++
	}

	fmt.Printf("finished in scene %q after %d steps (%.1f game seconds)\n",
		obs.Scene, steps, float64(steps)/60)
}
