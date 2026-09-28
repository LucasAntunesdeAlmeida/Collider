# AGENTS.md

You are reading the repo of Collider, a code-first 2D game engine for Go
whose games are playable by AI agents out of the box. That includes you:
any example here can be run headless and driven over MCP or a step loop,
right now, from this checkout. This file covers both jobs: working on
the engine, and playing the games.

## Repo map

| Path | What it is |
|------|------------|
| `*.go` (root) | The engine: one flat package, `collider` |
| `agent.go` | Agent play: `Headless`, `Step`, `Observe`, `Autopilot`, MCP entry |
| `internal/mcps/` | Minimal MCP server (stdio, JSON-RPC 2.0), no dependencies |
| `examples/` | Fifteen complete games, one folder each, each with README and demo GIF |
| `examples/agent/` | Gem Rush: the reference agent-played game (no human input at all) |
| `tools/genassets/` | Regenerates all placeholder art from ASCII pixel grids |
| `tools/cover/` | Regenerates `.github/social-preview.png` from real assets |
| `PUBLISHING.md` | Shipping games as .exe or browser WebAssembly |

## Build, test, verify

```bash
go build ./...
go test ./...            # unit + headless feature tests, no window needed
```

CI (`.github/workflows/ci.yml`) enforces, in order: `gofmt -l .` clean,
`go vet ./...`, `staticcheck ./...`, `go test ./...`. Run all four
before considering a change done. On Linux, tests need a display, so CI
wraps them in `xvfb-run`; do the same in any Linux sandbox.

Examples run from their own folder so asset paths resolve:

```bash
cd examples/pong && go run .
```

Placeholder art is generated, never hand-edited: change the ASCII grids
in `tools/genassets` and run `go run ./tools/genassets` from the root.

## Engine conventions

- The public API is deliberately small and fully listed in the README
  cheat sheet. If a call is not there, it does not exist. Adding API is
  a design decision, not a convenience: prefer composing what exists.
- Every feature is proven by an example game. If you add engine
  behavior, an example must exercise it.
- The README section "Behavior you can rely on" is a contract
  (Destroy is always deferred, collisions fire on enter, text never
  collides, Restart semantics). Do not break it; tests cover it.
- Examples stay in one small file with minimal boilerplate. That is the
  product: if an example grows scaffolding, the engine is missing
  something.

## Playing a game (yes, you)

Every Collider game exposes the same three-verb interface. State comes
back as one observation per step:

```json
{"scene": "play", "objects": [
  {"tag": "player", "x": 400, "y": 520, "w": 48, "h": 48},
  {"tag": "comet", "x": 312, "y": 180, "vy": 260, "w": 24, "h": 24},
  {"text": "SCORE 12", "x": 64, "y": 24, "w": 96, "h": 16}
]}
```

Objects carry `tag`, position (`x`, `y` is the center), velocity
(`vx`, `vy`), collision size (`w`, `h`: the hitbox, which can be
smaller than the sprite), `solid`, and `text` for HUD labels, so
scores and instructions are readable. The player is tagged `"player"`.
Decorative scenery is filtered out for you. Positions are world
coordinates; HUD objects pinned to the screen report screen coordinates
and carry `"fixed": true`. While an overlay (a pause menu) is open,
`"overlay"` names it: it takes all input, and the scene under it is
frozen; its objects are listed first, then the overlay's.

**Over MCP** (any client, or you via a subprocess): run a game with the
`COLLIDER_AGENT=mcp` env var and its stdio becomes an MCP server with
three tools. With `COLLIDER_AGENT=mcp-window` the same server runs
while the game plays windowed in real time: agent keys merge with the
human keyboard, `act`'s input stays held between calls like a
controller, and games that set `AgentDocs` return their rules in the
initialize instructions.

- `observe` returns the current observation (plus a `state` field when
  the game attached one with `AgentState`).
- `act` holds input and advances time: `keys` is any keyboard key by
  name (`"J"`, `"Numpad1"`, `"ShiftLeft"`; aliases `Left, Right, Up,
  Down, Space, Enter, Esc`), or, when the game declared
  `g.Controls(...)`, the control names the tool description lists.
  Optional `click` at `x`, `y` (a press on the first frame), `down` to
  hold the pointer at `x`, `y` for every frame (drags, virtual
  joysticks), and `frames` for how many 1/60s frames to hold (default
  10). Returns the resulting observation; unknown key
  names return an error naming the valid controls.
- `reset` restarts a scene by name.

The protocol is newline-delimited JSON-RPC 2.0 (`initialize`,
`tools/list`, `tools/call`), so a plain subprocess pipe works when no
MCP client is handy.

**In process** (bots, CI, playtests):

```go
g.Headless("play")
obs := g.Step(engine.Action{Keys: []engine.Key{engine.Right}})
```

Fixed 60 steps per second, fully deterministic: same actions, same
results. `examples/agent/bot` is a complete bot that wins Gem Rush this
way and prints how long it took.

**Watching**: `g.Autopilot(fn)` runs the game windowed with your
function supplying input each frame. Set `COLLIDER_RECORD=demo.gif` and
the session saves as a GIF: this is how an agent records a demo of its
own play (`examples/agent/demo.gif` was made exactly like that).

Games can opt out of all of this with `g.DisallowAgents()`.

## Shipping a Collider game? Tell agents it is playable

If you are building a game with this engine, put this in your own
repo's AGENTS.md so visiting agents know they can play it:

```markdown
This game is built with Collider and is agent-playable. Run it with
COLLIDER_AGENT=mcp and it serves MCP on stdio (observe, act, reset).
The player object is tagged "player". See
https://github.com/LucasAntunesdeAlmeida/collider for the protocol.
```
