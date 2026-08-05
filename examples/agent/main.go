// Gem Rush: a game with NO human input. Running it opens a window where
// the built-in agent plays, using only observations, exactly like an
// external agent would.
//
//	go run .                        watch the agent play
//	$env:COLLIDER_AGENT="mcp"       serve MCP on stdio for external agents
//	cd bot && go run .              headless run, no window (CI style)
//
// With COLLIDER_RECORD set, the agent records its own demo GIF and the
// program ends itself after ~22 seconds.
package main

import (
	"os"

	engine "github.com/LucasAntunesdeAlmeida/collider"
	"github.com/LucasAntunesdeAlmeida/collider/examples/agent/game"
)

func main() {
	g := game.New()

	if os.Getenv("COLLIDER_AGENT") != "" {
		g.Run("play") // Run serves MCP itself when the env var is set
		return
	}

	recording := os.Getenv("COLLIDER_RECORD") != ""
	frames := 0
	g.Autopilot(func(obs engine.Observation) engine.Action {
		frames++
		if recording && frames > 60*22 {
			g.Quit() // a recording session ends itself
		}
		return game.Decide(obs)
	})
	g.Run("play")
}
