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

import "github.com/LucasAntunesdeAlmeida/collider/examples/agent/game"

func main() {
	game.New().Run("play")
}
