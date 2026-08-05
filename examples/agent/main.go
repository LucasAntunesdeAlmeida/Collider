// Gem Rush, human edition: WASD or arrows, collect all gems in time.
//
// The same binary is agent-playable, because agent play is on by
// default in Collider:
//
//	$env:COLLIDER_AGENT="mcp"; go run .
//
// runs it headless as an MCP server on stdio: connect any MCP client
// (Claude, for example) and it can observe, act and reset.
//
// See bot/ for a plain Go program playing it through the same API.
package main

import (
	"os"

	"github.com/LucasAntunesdeAlmeida/collider/examples/agent/game"
)

func main() {
	// Humans get the menu; agents jump straight into the round.
	scene := "menu"
	if os.Getenv("COLLIDER_AGENT") != "" {
		scene = "play"
	}
	game.New().Run(scene)
}
