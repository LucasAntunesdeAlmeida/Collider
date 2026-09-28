package collider

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/LucasAntunesdeAlmeida/collider/internal/mcps"
)

// Agent play: every Collider game can be played by a program or an AI
// agent, headless and deterministic. It is on by default and adds zero
// overhead until used; a game can opt out with DisallowAgents.
//
// Two ways in:
//   - In process: g.Headless("menu") then a g.Step(action) loop.
//   - Out of process: run the shipped game with COLLIDER_AGENT=mcp and
//     it serves observe/act/reset as MCP tools on stdio.

// Action is one frame of injected input.
type Action struct {
	Keys   []Key   `json:"-"` // keys held during the frame
	MouseX float64 `json:"mouseX"`
	MouseY float64 `json:"mouseY"`
	Click  bool    `json:"click"` // press the left button this frame
	// MouseDown holds the pointer at MouseX, MouseY this frame (drags,
	// virtual joysticks); Click implies it on its frame.
	MouseDown bool `json:"mouseDown"`
}

// Observation is the structured view of the current frame: the scene
// name, every object with its position, motion and label, and whatever
// extra state the game attached with AgentState. While an overlay is
// open (a pause menu), Overlay names it and Objects lists the frozen
// scene's objects first, then the overlay's.
type Observation struct {
	Scene   string      `json:"scene"`
	Overlay string      `json:"overlay,omitempty"`
	Objects []ObjectObs `json:"objects"`
	State   any         `json:"state,omitempty"`
}

// ObjectObs describes one live object. Tag is the game's own label
// (tag your player "player" so agents can find themselves); Text is
// set for text objects, so HUDs and scores are readable.
type ObjectObs struct {
	Tag   string  `json:"tag,omitempty"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Vx    float64 `json:"vx,omitempty"`
	Vy    float64 `json:"vy,omitempty"`
	W     float64 `json:"w"`
	H     float64 `json:"h"`
	Solid bool    `json:"solid,omitempty"`
	Text  string  `json:"text,omitempty"`
	// Fixed objects (the HUD) report screen coordinates; everything
	// else is in world coordinates, which differ once the scene moves
	// its camera.
	Fixed bool `json:"fixed,omitempty"`
}

// DisallowAgents turns agent play off for this game: Headless, Step and
// the MCP server refuse to run, and COLLIDER_AGENT is ignored. Agent
// play is allowed by default.
func (g *Game) DisallowAgents() {
	g.agentsOff = true
}

// agentCmd is one MCP command crossing into the windowed game loop.
type agentCmd struct {
	action  Action
	frames  int
	observe bool   // observation only: keep the currently held input
	reset   string // scene to restart first, "" = none
	reply   chan Observation
}

// AgentDocs sets the game's agent-facing documentation: rules, goals,
// coordinate conventions, anything an agent should know before playing.
// It is returned as the MCP server's instructions on initialize.
// Optional.
func (g *Game) AgentDocs(docs string) {
	g.agentDocs = docs
}

// Controls names this game's inputs for agents: action name to key,
// like {"jump": engine.Space, "p2-attack": ebiten.KeyNumpad1}. The MCP
// act tool then accepts these names and lists them in its description,
// so any agent discovers how to play without reading the game's source.
// Optional; raw key names always work.
func (g *Game) Controls(controls map[string]Key) {
	g.controls = controls
}

// AgentState attaches game-defined state to every observation: fn runs
// once per observation and its result appears as the "state" field.
// Use it for what object positions cannot express: score, health,
// phase, whose turn it is. Optional.
func (g *Game) AgentState(fn func() any) {
	g.stateFn = fn
}

// Headless prepares the game to be driven by Step instead of Run: no
// window, no audio, injected input, fixed 60 steps per second.
func (g *Game) Headless(scene string) {
	if g.agentsOff {
		panic("collider: agent play is disallowed by this game")
	}
	g.headless = true
	g.agentIn = &agentInput{keys: map[Key]bool{}}
	g.input = g.agentIn
	g.Go(scene)
	g.advance(0) // activate the scene so Observe works immediately
}

// Autopilot runs the game windowed while an agent function supplies the
// input: fn receives each frame's observation and returns the action to
// hold. Watch a bot play, or combine with COLLIDER_RECORD and let the
// agent record its own demo GIF. Call before Run.
func (g *Game) Autopilot(fn func(Observation) Action) {
	if g.agentsOff {
		panic("collider: agent play is disallowed by this game")
	}
	g.agentIn = &agentInput{keys: map[Key]bool{}}
	g.input = g.agentIn
	g.pilot = fn
}

func (g *Game) injectAction(a Action) {
	in := g.agentIn
	clear(in.keys)
	for _, k := range a.Keys {
		in.keys[k] = true
	}
	in.x, in.y = a.MouseX, a.MouseY
	in.click = a.Click
	in.down = a.MouseDown
}

// Step advances exactly one frame with the given input and returns the
// resulting observation. Deterministic: same actions, same results.
func (g *Game) Step(a Action) Observation {
	if !g.headless {
		panic("collider: call Headless before Step")
	}
	g.injectAction(a)
	g.advance(1.0 / 60.0)
	g.agentIn.click = false
	return g.Observe()
}

// Observe returns the structured state of the current frame without
// advancing it.
func (g *Game) Observe() Observation {
	obs := Observation{}
	if g.current == nil {
		return obs
	}
	obs.Scene = g.current.name
	obs.Objects = g.current.observe(obs.Objects)
	if g.overlay != nil {
		obs.Overlay = g.overlay.name
		obs.Objects = g.overlay.observe(obs.Objects)
	}
	if g.stateFn != nil {
		obs.State = g.stateFn()
	}
	return obs
}

// observe appends the scene's live objects as agents see them.
func (s *Scene) observe(out []ObjectObs) []ObjectObs {
	for _, o := range s.objects {
		if o.dead || (o.visual && o.textStr == "") {
			continue // scenery is noise for agents; text still matters
		}
		out = append(out, ObjectObs{
			Tag: o.tag, X: o.X, Y: o.Y, Vx: o.Vx, Vy: o.Vy,
			W: o.w, H: o.h, Solid: o.solid, Text: o.textStr, Fixed: o.fixed,
		})
	}
	return out
}

// controlNames lists declared control names, sorted for stable output.
func (g *Game) controlNames() []string {
	names := make([]string, 0, len(g.controls))
	for n := range g.controls {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// resolveKeys turns agent-supplied key names (declared control names or
// keyboard key names) into keys, erroring on anything unknown so agents
// get corrected instead of silently ignored.
func (g *Game) resolveKeys(names []any) ([]Key, error) {
	var keys []Key
	for _, n := range names {
		name, ok := n.(string)
		if !ok {
			return nil, fmt.Errorf("keys must be strings, got %v", n)
		}
		if k, ok := g.controls[name]; ok {
			keys = append(keys, k)
			continue
		}
		if k, ok := keyNames[name]; ok {
			keys = append(keys, k)
			continue
		}
		hint := `a keyboard key name like "J", "Numpad1" or "ArrowLeft"`
		if len(g.controls) > 0 {
			hint = "one of this game's controls [" +
				strings.Join(g.controlNames(), ", ") + "] or " + hint
		}
		return nil, fmt.Errorf("unknown key %q: use %s", name, hint)
	}
	return keys, nil
}

// serveMCP exposes the game as an MCP server on stdio: observe, act
// (with a frames count, since agents think slower than 60fps) and
// reset. Headless, tools run in place and step the simulation; in
// windowed mode (COLLIDER_AGENT=mcp-window) they cross to the game
// loop through the command queue and real time keeps flowing.
func (g *Game) serveMCP() {
	windowed := g.agentReq != nil
	obsJSON := func() string {
		b, _ := json.Marshal(g.Observe())
		return string(b)
	}
	// send runs a command on the game thread and waits for its
	// observation; used only in windowed mode. If the window closes
	// while a call is in flight, it answers with the final state
	// instead of hanging the client.
	send := func(cmd agentCmd) string {
		cmd.reply = make(chan Observation, 1)
		select {
		case g.agentReq <- cmd:
		case <-g.agentDone:
			return obsJSON()
		}
		select {
		case obs := <-cmd.reply:
			b, _ := json.Marshal(obs)
			return string(b)
		case <-g.agentDone:
			return obsJSON()
		}
	}

	// The act tool describes this game's own controls when declared,
	// so the server is self-documenting for any MCP client.
	keysHelp := `any keyboard key name, e.g. "J", "Numpad1", "ShiftLeft" (aliases: Left, Right, Up, Down, Space, Enter, Esc)`
	if len(g.controls) > 0 {
		keysHelp = "this game's controls: [" + strings.Join(g.controlNames(), ", ") +
			"] (raw keyboard key names also work)"
	}
	obsHelp := "Get the current game state: scene name and all objects (tag, position, velocity, size, text). The player object is usually tagged \"player\". While an overlay (a pause menu) is open, \"overlay\" names it: it takes the input, and the scene under it is frozen."
	if g.stateFn != nil {
		obsHelp += " The \"state\" field carries game-specific state."
	}

	mcps.Serve(g.title, g.agentDocs, []mcps.Tool{
		{
			Name:        "observe",
			Description: obsHelp,
			Schema:      `{"type":"object","properties":{}}`,
			Call: func(args map[string]any) (string, error) {
				if windowed {
					return send(agentCmd{observe: true, frames: 1}), nil
				}
				return obsJSON(), nil
			},
		},
		{
			Name:        "act",
			Description: "Hold keys and/or click, then advance the game. keys: " + keysHelp + ". frames: how many 1/60s frames to advance with this input held (default 10). click presses the mouse at x,y on the first frame. down holds the pointer (mouse button or finger) at x,y for all the frames: drags and virtual joysticks.",
			Schema:      `{"type":"object","properties":{"keys":{"type":"array","items":{"type":"string"}},"frames":{"type":"number"},"click":{"type":"boolean"},"down":{"type":"boolean"},"x":{"type":"number"},"y":{"type":"number"}}}`,
			Call: func(args map[string]any) (string, error) {
				a := Action{}
				if ks, ok := args["keys"].([]any); ok {
					keys, err := g.resolveKeys(ks)
					if err != nil {
						return "", err
					}
					a.Keys = keys
				}
				if x, ok := args["x"].(float64); ok {
					a.MouseX = x
				}
				if y, ok := args["y"].(float64); ok {
					a.MouseY = y
				}
				a.Click, _ = args["click"].(bool)
				a.MouseDown, _ = args["down"].(bool)
				frames := 10
				if f, ok := args["frames"].(float64); ok && f >= 1 {
					frames = int(f)
				}
				if windowed {
					// Real time: the input is held for the frames,
					// then the observation comes back. It stays held
					// after, like a controller, until the next act.
					return send(agentCmd{action: a, frames: frames}), nil
				}
				for range frames {
					g.Step(a)
					a.Click = false
				}
				return obsJSON(), nil
			},
		},
		{
			Name:        "reset",
			Description: "Restart a scene to its initial state and switch to it.",
			Schema:      `{"type":"object","properties":{"scene":{"type":"string"}},"required":["scene"]}`,
			Call: func(args map[string]any) (string, error) {
				scene, _ := args["scene"].(string)
				if windowed {
					return send(agentCmd{reset: scene, observe: true, frames: 1}), nil
				}
				g.Restart(scene)
				g.advance(0)
				return obsJSON(), nil
			},
		},
	})
}
