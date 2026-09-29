# Collider

**The fastest way to a playable 2D game in Go. Every game you build can be
played by humans and by AI agents, out of the box.**

![Collider: 2D games in Go, played by humans and AI agents](.github/social-preview.png)

Collider is a code-first 2D game engine built around one idea: games are made of
objects that collide, and things that happen when they do. You import the library,
describe your objects, attach events, and you have a game. No editor, no project
files, no boilerplate game loop.

```go
player.OnCollisionWith("enemy", func(e *engine.Object) {
    g.Sound("hit.wav")
    g.Go("gameover")
})
```

That covers the humans. For the agents, the very same binary you ship to
players is also a deterministic headless environment for bots and CI, and
an MCP server that any AI agent can connect to and play, with zero extra
code in your game:

```powershell
claude mcp add mygame -e COLLIDER_AGENT=mcp -- C:\path\to\mygame.exe
```

![Zombie Night, one of the example games](examples/zombie-night/demo.gif)

**What's in the box:**

- Collision-first core: enter-only collision events, tag rules, solid
  resolution, gravity, spatial hash broad phase
- **AI agents can play every game made with Collider**, out of the box:
  headless and deterministic for bots and CI, over MCP for any AI agent
  (Claude, Cursor, or a plain script), or on-screen via autopilot; one
  example game is played by its own agent and even records its demo GIF
  itself
- Scenes, timers, sprite-sheet animations, text with custom TTF fonts
  (with fallback fonts for any script, Chinese included), sound and
  music (wav/ogg)
- Fifteen example games across genres, from pong to a dungeon crawler,
  each one a template you can start from
- Shipping built in: embed assets into a single .exe or build for the
  browser; one example is [live on itch.io](https://lucasantunesdealmeida.itch.io/ship)
- A GIF recorder for demos, and generated placeholder assets so nothing
  here needs hand-drawn art to run

---

## Design principles

1. **Fewest lines possible.** A complete game with menus, music, sprites and
   logic should fit in one small file. Every line of required boilerplate is a bug.
2. **Collisions are the core.** Clicks are point-vs-object collisions. Triggers,
   solid walls, bullets hitting enemies: one unified system with events.
3. **Sensible defaults.** Sprites collide with their image bounds. Sounds just
   play. Scenes just switch. Configuration is always optional.
4. **Scenes are everything.** A menu, a level and a game-over screen are all the
   same concept. Learn one thing, build every screen of your game with it.
5. **No magic.** Plain Go, plain callbacks, `go get` and read the code.

Built on top of [Ebitengine](https://ebitengine.org/) for windowing, rendering,
input and audio. Collider is the game-object, collision and event layer that
every Ebitengine project currently rebuilds by hand.

---

## Install

```bash
go get github.com/LucasAntunesdeAlmeida/collider
```

```go
import engine "github.com/LucasAntunesdeAlmeida/collider"
```

---

## Hello, collision

The smallest complete program: two objects, one event.

```bash
go run ./examples/hello
```

```go
package main

import (
	"fmt"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

func main() {
	g := engine.New("Hello, collision", 800, 600)
	play := g.Scene("play")

	box := play.Add(engine.Rect(64, 64, engine.Red).At(400, 300))
	player := play.Add(engine.Rect(48, 48, engine.Blue).At(100, 300))

	player.OnUpdate(func(dt float64) {
		if g.Key(engine.Right) {
			player.Move(200*dt, 0)
		}
		if g.Key(engine.Left) {
			player.Move(-200*dt, 0)
		}
	})

	player.OnCollision(func(other *engine.Object) {
		fmt.Println("hit!")
	})

	box.OnClick(func() {
		box.Destroy()
	})

	g.Run("play")
}
```

---

## Example games

Fifteen complete games, each exercising a different part of the engine
and each a starting point for your own. Run any of them from its own
folder so the asset paths resolve:

```bash
cd examples/pong
go run .
```

| Game | Genre | What it proves |
|------|-------|----------------|
| [Hello](examples/hello/) | Smallest program | Objects, movement, collision and click events |
| [Pong](examples/pong/) | Arcade / versus | Built-in velocity, solid bounce, score text |
| [Jumper](examples/jumper/) | Platformer | Gravity, solid ground, `Grounded()`, pickups |
| [Zombie Night](examples/zombie-night/) | Top-down shooter | Runtime spawning, tags, scene collision rules, timers |
| [Horde](examples/horde/) | Survivor | A world bigger than the window: camera follow, screen-fixed HUD, draw layers; drawing effects (flip, rotate, tint, flash, fade); area queries (auto-aim, contact damage); hitboxes smaller than the sprites; an overlay pause menu with number-key shortcuts (keys by name); a drag joystick and pause on focus loss; a mute key and separate music and sound volumes; full gamepad play with pad prompts; a QUIT entry where quitting is possible |
| [Breakout](examples/breakout/) | Brick breaker | Grid spawning, win conditions, mouse control, menus |
| [Memory](examples/memory/) | Point-and-click puzzle | A game with zero movement, pure click events |
| [Arena](examples/arena/) | Survival | Extending the engine: your own types embedding `Object` with custom methods |
| [Caves](examples/caves/) | Exploration | Procedural generation: cellular automata map, BFS-guaranteed winnable |
| [Runner](examples/runner/) | Endless runner | Sprite sheet animations (run/jump/death), moving-world auto-scroll |
| [Dialog](examples/dialog/) | Visual novel | Custom TTF fonts, text size and color, typewriter effect; speaks the player's language (`g.Language`), Chinese included, through a CJK fallback font |
| [Dungeon](examples/dungeon/) | Dungeon crawler | Multiple screens: ASCII room layouts, edge transitions, key and lock |
| [Catcher](examples/catcher/) | Arcade | Save data that persists on desktop and in the browser: `g.Save` / `g.Load`; a record kept when the window closes mid-round: `g.OnClose` |
| [Ship](examples/ship/) | Arcade | Publishing: embedded assets, fullscreen/resizable/icon, exe and browser builds |
| [Gem Rush](examples/agent/) | Arcade | Agents only: no human input; played by its own pilot, headless bots, or MCP agents |

Each example folder has the game's code, an explanation and a demo GIF.
The placeholder assets are generated by `go run ./tools/genassets`; replace
them with real art and the games look like games.

## AI agents can play your game

Any game you build with Collider is agent-playable out of the box, with
zero overhead until used. An agent sees the game as structured data, one
observation per step: the scene, and every object with its tag, position,
motion, size and text (so HUDs and scores are readable):

```json
{"scene": "play", "objects": [
  {"tag": "player", "x": 400, "y": 520, "w": 48, "h": 48},
  {"tag": "comet", "x": 312, "y": 180, "vy": 260, "w": 24, "h": 24},
  {"text": "SCORE 12", "x": 64, "y": 24, "w": 96, "h": 16}
], "state": {"lives": 2, "wave": 3}}
```

Positions are world coordinates; objects pinned to the screen with
`Fixed` (a HUD) report screen coordinates and carry `"fixed": true`.
`w` and `h` are the collision box: an object's `Hitbox` when it has
one, so agents dodge exactly what can hit them.
While an overlay is open (a pause menu), `"overlay"` names it: it takes
all input, and `objects` lists the frozen scene's objects, then the
overlay's. Once the game calls `Quit` (its QUIT entry), `"quit": true`
marks the run as over.

Three optional calls make any game, however complex, fully
agent-playable and self-describing:

- `g.Controls(map[string]engine.Key{"jump": engine.Space, ...})` names
  your inputs; the MCP act tool accepts these names and lists them in
  its own description, so agents discover how to play on connect.
- `g.AgentState(func() any {...})` attaches game state (score, health,
  phase) to every observation as the `state` field above.
- `g.AgentDocs("...")` serves your game's rules to agents when they
  connect (the MCP initialize instructions).

Four ways in:

- **Headless** (bots, CI): `g.Headless("play")` then a loop of
  `g.Step(engine.Action{Keys: ...})`, each step returning the next
  observation. Deterministic: same actions, same results. Playtest your
  game with a bot on every commit.
- **MCP** (AI agents): run any Collider game with `COLLIDER_AGENT=mcp`
  and it serves MCP on stdio (`observe`, `act`, `reset` tools). Any MCP
  client can connect and play your game; `act` accepts any keyboard key
  by name, or the names you declared with `Controls`, plus `click` or
  a held pointer (`down`) at `x`, `y`, and gamepad buttons (`pad`)
  and the left stick (`stickX`, `stickY`) for games played with a pad.
- **Windowed MCP** (fight the AI): `COLLIDER_AGENT=mcp-window` opens
  the normal window running in real time while serving the same MCP
  tools. Agent input merges with the keyboard, so a person and an
  agent can play the same game together, and COLLIDER_RECORD captures
  the match.
- **Autopilot** (watch it): `g.Autopilot(fn)` runs the game windowed
  while your agent function supplies the input each frame. Combine with
  the GIF recorder and an agent records your demo for you.

### Connect any MCP client in one command

The game itself is the MCP server (stdio): there is nothing to install.
With Claude:

```powershell
claude mcp add mygame -e COLLIDER_AGENT=mcp -- C:\path\to\mygame.exe
```

Cursor, Windsurf, VS Code and every other MCP client take the same
shape, a command plus one env var:

```json
{"mcpServers": {"mygame": {
  "command": "C:/path/to/mygame.exe",
  "env": {"COLLIDER_AGENT": "mcp"}
}}}
```

The protocol is plain newline-delimited JSON-RPC, so even a Python
script can drive a game through a subprocess.

Tag your player `"player"` so agents can find themselves. A game can opt
out entirely with `g.DisallowAgents()`, which also makes it ignore
`COLLIDER_RECORD`. The
[Gem Rush example](examples/agent/) is a game with no human input at
all: its own pilot plays it, and its demo GIF is agent-recorded.

## Publish your game

Collider games ship as a single self-contained .exe (assets embedded via
`engine.UseAssets` and `go:embed`) or run in the browser via WebAssembly,
ready for itch.io. [PUBLISHING.md](PUBLISHING.md) is the complete guide;
the [ship example](examples/ship/) has it all wired up, and it is
actually published: **[play Ship in your browser on
itch.io](https://lucasantunesdealmeida.itch.io/ship)**.

## Performance

A frame's work scales with what is on screen and what actually
touches, not with bookkeeping. The engine's own benchmark
(`go test -bench Scene -benchmem`) runs a survivor-style scene of 1000
live objects: 400 enemies crowding the hero, 150 projectiles, 300
visual gems and 150 props and walls, with the camera following. On a
Ryzen 7 5800H one frame of logic (motion, broad phase, collision
events, solids) takes about **0.75 ms with zero allocations**, and
preparing the draw about 0.26 ms, since objects outside the view are
skipped. That leaves nearly all of a 60 fps frame to your game.

## Record a GIF of your game

Every Collider game can record itself, no tooling needed. Set the
`COLLIDER_RECORD` environment variable to a file path, play, close the
window, and the session is saved as an animated GIF (the demos above were
made exactly this way):

```powershell
$env:COLLIDER_RECORD="demo.gif"; go run .
```

A game that calls `g.DisallowAgents()` ignores `COLLIDER_RECORD`, so a
shipped build never writes a GIF wherever its environment says; call it
only in release builds (a build tag) to keep recording demos.

---

# API cheat sheet

The complete public surface implied by the example games. If it is not here, it
does not exist. That is the point.

### Game

| Call | Meaning |
|------|---------|
| `engine.New(title, w, h) *Game` | Create the game/window |
| `g.Scene(name) *Scene` | Create (or fetch) a scene |
| `g.Go(name)` | Switch scenes (state preserved) |
| `g.Restart(name)` | Switch scenes, resetting it to its initial state |
| `g.Overlay(name)` | Show a scene on top (pause, menus): it runs, the scene under it draws frozen |
| `g.CloseOverlay()` | Remove the overlay; the scene under it resumes where it froze |
| `g.Run(name)` | Start the loop on a scene (blocks) |
| `g.Key(k) bool` | Is this key held? |
| `engine.KeyNamed(name) Key` | Any key by name, for keys without a constant: `"1"`/`"Digit1"` (number row), `"Numpad1"`, `"J"`, `"F1"`, `"Tab"`, `"ShiftLeft"`...; the names agents use; panics on an unknown name |
| `g.Mouse() (x, y)` | Cursor position (a finger on a touch screen counts) |
| `g.MouseDown() bool` | Is the left button (or any finger) held? Drags, virtual joysticks |
| `g.Focused() bool` | Does the window have focus? Pause when the player switches away (always true for agents) |
| `g.PadDown(b) bool` | Is this gamepad button held on any connected pad? `engine.PadA`, `PadB`, `PadX`, `PadY`, `PadLB`, `PadRB`, `PadLT`, `PadRT` (triggers pulled), `PadBack`, `PadStart`, `PadLStick`, `PadRStick` (sticks pressed in), `PadUp`, `PadDown`, `PadLeft`, `PadRight` (d-pad); standard layout by position: Xbox, PlayStation, Switch Pro, Steam Deck |
| `g.PadAxis(a) float64` | A stick, raw -1..1 with Y down: `engine.PadLeftX`, `PadLeftY`, `PadRightX`, `PadRightY`; no dead zone (ignore small tilts yourself); the pad pushed furthest answers |
| `g.PadConnected() bool` | Is a gamepad connected? Switch on-screen prompts (agents: from their first pad action on) |
| `g.Sound(path)` | Fire-and-forget sound effect (silent for agents and tests, see below) |
| `g.Volume(v)` | Master volume 0..1 (0 mutes) for the playing music and every sound started after; scales both channels below |
| `g.MusicVolume(v)` | Music volume 0..1, times the master: applies to the playing music at once (a settings slider) |
| `g.SoundVolume(v)` | Sound effect volume 0..1, times the master, for every `Sound` started after (0: they do no work) |
| `g.Save(key, v) error` | Persist a JSON-encodable value (file on desktop, localStorage in the browser, memory for agents and tests) |
| `g.Load(key, &v) bool` | Read it back; false (v untouched) when missing or undecodable |
| `g.Quit()` | End the game after this frame: the window closes and `Run` returns (does nothing in a browser; headless, the run ends) |
| `g.CanQuit() bool` | Can Quit end the game here? False in a browser, where a page cannot close its tab: hide the QUIT entry |
| `g.Language() string` | The player's language as a BCP 47 tag (`"pt-BR"`, `"zh-CN"`, `"en-US"`), `""` when unknown: pick your translation at startup. `""` for agents and tests; `COLLIDER_LANG=pt-BR` overrides it to try one |
| `g.OnClose(fn)` | Run fn once when the player closes the window (X, Alt+F4, a launcher's exit), then end the game like Quit: save or bank progress there. Never runs headless, in a browser, or on Quit |
| `g.Fullscreen(on)` / `g.Resizable(on)` / `g.Icon(path)` | Window polish |
| `g.Headless(scene)` / `g.Step(action)` / `g.Observe()` | Agent play, headless and deterministic |
| `g.Controls(map[string]Key)` | Name your inputs so agents can discover them |
| `g.AgentState(fn)` | Attach game state to every observation |
| `g.AgentDocs(text)` | Game rules served to agents on MCP connect |
| `g.DisallowAgents()` | Lock a shipped game: no agent play (on by default), `COLLIDER_AGENT` and `COLLIDER_RECORD` ignored |

### Scene

| Call | Meaning |
|------|---------|
| `s.Add(obj) *Object` | Put an object in the scene (any time, even mid-game) |
| `s.Music(path)` | Looping background music while the scene is active (silent for agents and tests) |
| `s.Gravity(px_per_s2)` | Gravity for `WithGravity()` objects |
| `s.Every(sec, fn)` | Repeating timer |
| `s.After(sec, fn)` | One-shot timer |
| `s.OnClick(fn(x, y))` | Click anywhere in the scene |
| `s.OnUpdate(fn(dt))` | Runs every frame while the scene is active |
| `s.OnCollision(tagA, tagB, fn(a, b))` | Rule for every current and future pair |
| `s.Count(tag) int` | How many objects with this tag are alive |
| `s.Camera(x, y)` | Center the view on a world point (follow the player every frame) |
| `s.Near(tag, x, y, r) []*Object` | Live objects with this tag whose box overlaps the circle, closest first (visual included, Fixed not) |

### Object: constructors and configuration (chainable)

| Call | Meaning |
|------|---------|
| `engine.Sprite(path)` | Object from an image (collider = image bounds) |
| `engine.Rect(w, h, color)` | Colored rectangle object |
| `engine.Text(str)` | Text object (clickable, but never collides) |
| `engine.UseAssets(fs)` | Load all assets from an embedded filesystem (go:embed) |
| `.At(x, y)` | Position (center); `.AtEdge()` picks a random screen edge |
| `.Tag(name)` | Label for tag-based collision rules |
| `.Solid()` | Engine resolves overlaps (walls, floors, paddles) |
| `.WithGravity()` | Affected by the scene's gravity |
| `.Visual()` | Decoration: drawn, never collides |
| `.Fixed()` | Pinned to the screen: ignores the camera, never collides (HUD) |
| `.Layer(n)` | Draw order: lower layers first, insertion order within a layer; clicks hit the top |
| `.Size(w, h)` | Override the drawn size (also the collider, unless Hitbox is set) |
| `.Hitbox(w, h)` | Collision box apart from the drawn size, centered: collisions, solids, `Touching`, `Near` and agents' `w`/`h` use it; drawing and clicks keep the drawn bounds |
| `.Animation(name, strip, frames, fps)` | Define a sprite-sheet animation |
| `.Font(paths...)` | Custom TTF/OTF font. More paths are fallbacks: the first font draws every glyph it has, the next ones fill in the rest (a CJK font behind a Latin pixel font), each loaded the first time a text needs it |
| `.TextSize(px)` / `.TextColor(c)` | Text styling (size, color) |

### Object: runtime

| Call | Meaning |
|------|---------|
| `o.X, o.Y, o.Vx, o.Vy` | Position and velocity; velocity applies every frame |
| `o.Data` | Free `any` field for your game state |
| `o.Move(dx, dy)` | Move (respects solid collisions) |
| `o.MoveToward(x, y, dist)` / `o.VelocityToward(x, y, speed)` | Homing helpers |
| `o.Grounded() bool` | Resting on a solid (platformers) |
| `o.SetText(s)` / `o.SetSprite(path)` | Change content at runtime |
| `o.Play(name)` / `o.PlayOnce(name)` | Switch animations (loop / hold last frame) |
| `.Alpha(a)` | Opacity, 0 (invisible) to 1 (opaque, the default): fades, ghosts |
| `.Tint(c)` | Multiply the colors by c (nil clears): color variants of one sprite |
| `o.Flash(c, sec)` | Draw a solid silhouette of color c for sec seconds (hit flash) |
| `.Rotate(rad)` | Draw turned around the center, clockwise; the collider stays upright |
| `.FlipX(on)` | Draw mirrored horizontally (face the way you walk); drawing only |
| `o.LifeTime(sec)` | Auto-destroy after n seconds |
| `o.Destroy()` | Remove from scene (safe inside callbacks; applied at end of frame) |
| `o.OnUpdate(fn(dt))` | Per-frame logic |
| `o.OnClick(fn)` | Clicked (point-vs-bounds collision) |
| `o.OnCollision(fn(other))` | Touching anything |
| `o.OnCollisionWith(tag, fn(other))` | Touching anything with this tag |
| `o.Touching(tag) []*Object` | What overlaps o right now, with this tag (continuous contact; never visual) |

---

## Behavior you can rely on

1. **`Destroy()` is always safe**, including inside any callback: the object
   is removed at the end of the frame, never mid-iteration.
2. **Collision events fire on *enter*.** `OnCollision` fires once when contact
   begins, not on every frame of overlap. Continuous contact (standing on the
   floor) is one event. To act on every frame of contact (damage over
   time), ask `o.Touching(tag)` instead.
3. **Solid vs trigger is the whole physics story.** `Solid()` objects push
   others out; everything else just reports contact. No forces, no torque,
   no restitution. If your game needs Angry Birds physics, this is not its
   engine.
4. **Text never collides.** A score label overlapping the ball will not
   bounce it. Text objects render and take clicks, nothing else.
5. **`Restart` rolls a scene back to its setup.** Objects added before the
   scene first runs are restored (even if destroyed); objects and timers
   spawned during play are dropped. Variables captured in your closures are
   yours to reset.
6. **A tap is a click.** `g.Mouse()` and every `OnClick` read touch as well
   as the mouse, so a game built for a mouse works on a phone without a
   second input path. A lifted finger leaves the cursor where it was, so
   anything following the cursor stays put until the next touch.
   `g.MouseDown()` is true while the button or any finger is held.
7. **An overlay freezes the scene under it.** While `g.Overlay` is open,
   only the overlay updates, ticks timers, takes clicks and collides;
   the scene under it still draws but does not advance at all. Opening
   and closing apply at the start of the next frame, like `g.Go`, and
   `Go`/`Restart` close any overlay.
8. **`Quit` ends the game after the frame.** The frame that calls
   `g.Quit()` finishes (its updates, deferred destroys and draw), then
   the window closes and `Run` returns. In a browser a page cannot
   close its tab, so `Quit` does nothing there and `g.CanQuit()` is
   false. Headless, the run ends instead: every later `Step` advances
   nothing and returns the final observation with `quit: true`, until
   `Headless` (or the MCP `reset` tool) starts a new run. Closing the
   window runs the `g.OnClose` handler once (if the game set one)
   between two frames, then ends the game the same way; no frame runs
   after it.
9. **Audio never gets in the way.** Agent and test runs (`Headless`,
   `Autopilot`, `COLLIDER_AGENT=mcp`, `go test`) are silent: `Sound`
   and `Music` do no audio work and the sound device is never opened.
   These are the same runs whose saves stay in memory. A
   `COLLIDER_AGENT=mcp-window` session is a real window a person
   watches and plays along with, so it has sound. A machine with no
   working audio device (a Linux box or container without ALSA,
   PulseAudio or PipeWire) plays silently after one log line, instead
   of failing to start.
10. **Agents and tests get the default language.** `g.Language()` is
   `""` in the same agent and test runs (Headless and Autopilot once
   called, `COLLIDER_AGENT=mcp`, `go test`), so a translated game
   plays in its default language and agents read the same text on
   every machine. The `COLLIDER_LANG` environment variable (`pt-BR`,
   `zh-CN`; `C` for none) overrides the answer everywhere, agent runs
   included, to try or test a translation.
11. **A fallback font never moves text it does not draw.** With
   `.Font(first, fallback)`, text the first font draws entirely
   measures and draws exactly as with that font alone. Glyphs from a
   fallback sit on the same baseline, and a line is as tall as the
   largest ascent plus the largest descent among the fonts it uses.

