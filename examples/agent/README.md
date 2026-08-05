# Gem Rush (built for agent play)

Collect all six gems before the timer runs out. Playable three ways,
which is the point of this example.

## 1. As a human

```bash
cd examples/agent
go run .
```

WASD or arrows. The game is ordinary Collider code; the only
agent-friendly choices are tags on everything that matters ("player",
"gem", "hud") and game state readable from a text object.

## 2. As a Go program (in process)

```bash
cd examples/agent/bot
go run .
```

The bot in [bot/main.go](bot/main.go) plays headless through the same
information an external agent gets: `g.Headless("play")`, then a loop of
`g.Step(action)` reading observations and chasing the nearest gem. No
window, no game internals, deterministic. It wins in about 5 game
seconds.

This is also the pattern for automated playtesting in CI: drive the
game, assert it reaches "win", no display needed.

## 3. As an AI agent (MCP)

```powershell
$env:COLLIDER_AGENT="mcp"; go run .
```

The same binary becomes an MCP server on stdio with three tools:
`observe` (scene and objects as JSON), `act` (hold keys, click, advance
N frames), `reset`. Connect any MCP client and it can play. Agent play
is on by default for every Collider game; a game can opt out with
`g.DisallowAgents()`.
