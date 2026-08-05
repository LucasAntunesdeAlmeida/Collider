package collider

import (
	"encoding/json"

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
}

// Observation is the structured view of the current frame: the scene
// name and every object with its position, motion and label.
type Observation struct {
	Scene   string      `json:"scene"`
	Objects []ObjectObs `json:"objects"`
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
}

// DisallowAgents turns agent play off for this game: Headless, Step and
// the MCP server refuse to run, and COLLIDER_AGENT is ignored. Agent
// play is allowed by default.
func (g *Game) DisallowAgents() {
	g.agentsOff = true
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

// Step advances exactly one frame with the given input and returns the
// resulting observation. Deterministic: same actions, same results.
func (g *Game) Step(a Action) Observation {
	if !g.headless {
		panic("collider: call Headless before Step")
	}
	in := g.agentIn
	clear(in.keys)
	for _, k := range a.Keys {
		in.keys[k] = true
	}
	in.x, in.y = a.MouseX, a.MouseY
	in.click = a.Click

	g.advance(1.0 / 60.0)

	in.click = false
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
	for _, o := range g.current.objects {
		if o.dead {
			continue
		}
		obs.Objects = append(obs.Objects, ObjectObs{
			Tag: o.tag, X: o.X, Y: o.Y, Vx: o.Vx, Vy: o.Vy,
			W: o.w, H: o.h, Solid: o.solid, Text: o.textStr,
		})
	}
	return obs
}

// serveMCP exposes the headless game as an MCP server on stdio:
// observe, act (with a frames count, since agents think slower than
// 60fps) and reset.
func (g *Game) serveMCP() {
	obsJSON := func() string {
		b, _ := json.Marshal(g.Observe())
		return string(b)
	}
	mcps.Serve(g.title, []mcps.Tool{
		{
			Name:        "observe",
			Description: "Get the current game state: scene name and all objects (tag, position, velocity, size, text). The player object is usually tagged \"player\".",
			Schema:      `{"type":"object","properties":{}}`,
			Call: func(args map[string]any) (string, error) {
				return obsJSON(), nil
			},
		},
		{
			Name:        "act",
			Description: "Hold keys and/or click, then advance the game. keys: array from [Left,Right,Up,Down,Space,Enter,Esc,W,A,S,D,F,R]. frames: how many 1/60s frames to advance with this input held (default 10). click presses the mouse at x,y on the first frame.",
			Schema:      `{"type":"object","properties":{"keys":{"type":"array","items":{"type":"string"}},"frames":{"type":"number"},"click":{"type":"boolean"},"x":{"type":"number"},"y":{"type":"number"}}}`,
			Call: func(args map[string]any) (string, error) {
				a := Action{}
				if ks, ok := args["keys"].([]any); ok {
					for _, kn := range ks {
						if k, ok := keyNames[kn.(string)]; ok {
							a.Keys = append(a.Keys, k)
						}
					}
				}
				if x, ok := args["x"].(float64); ok {
					a.MouseX = x
				}
				if y, ok := args["y"].(float64); ok {
					a.MouseY = y
				}
				a.Click, _ = args["click"].(bool)
				frames := 10
				if f, ok := args["frames"].(float64); ok && f >= 1 {
					frames = int(f)
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
				g.Restart(scene)
				g.advance(0)
				return obsJSON(), nil
			},
		},
	})
}
