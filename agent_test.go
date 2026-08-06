package collider

import (
	"strings"
	"testing"
)

// newAgentGame builds a tiny playable game: a player moved by arrow
// keys, a gem to touch, a button that switches scenes when clicked.
func newAgentGame() (*Game, *int) {
	g := New("agent-test", 800, 600)
	play := g.Scene("play")

	player := play.Add(Rect(20, 20, Blue).At(100, 300).Tag("player"))
	play.Add(Rect(20, 20, Yellow).At(300, 300).Tag("gem"))

	collected := 0
	player.OnUpdate(func(dt float64) {
		if g.Key(Right) {
			player.Move(400*dt, 0)
		}
		if g.Key(Left) {
			player.Move(-400*dt, 0)
		}
	})
	player.OnCollisionWith("gem", func(gem *Object) {
		gem.Destroy()
		collected++
	})

	btn := play.Add(Rect(100, 40, Green).At(700, 100).Tag("button"))
	btn.OnClick(func() { g.Go("menu") })
	g.Scene("menu").Add(Text("menu").At(400, 300))

	return g, &collected
}

func TestHeadlessStepMovesPlayer(t *testing.T) {
	g, _ := newAgentGame()
	g.Headless("play")

	obs := g.Observe()
	if obs.Scene != "play" {
		t.Fatalf("scene should be active after Headless, got %q", obs.Scene)
	}

	var before float64
	for _, o := range obs.Objects {
		if o.Tag == "player" {
			before = o.X
		}
	}

	var after float64
	for range 30 {
		obs = g.Step(Action{Keys: []Key{Right}})
	}
	for _, o := range obs.Objects {
		if o.Tag == "player" {
			after = o.X
		}
	}
	if after <= before {
		t.Fatalf("player should move right under injected input: %f -> %f", before, after)
	}
}

func TestHeadlessCollisionAndObservation(t *testing.T) {
	g, collected := newAgentGame()
	g.Headless("play")

	for range 120 {
		g.Step(Action{Keys: []Key{Right}})
	}
	if *collected != 1 {
		t.Fatalf("player should have collected the gem, got %d", *collected)
	}
	for _, o := range g.Observe().Objects {
		if o.Tag == "gem" {
			t.Fatal("collected gem should be gone from observations")
		}
	}
}

func TestHeadlessClick(t *testing.T) {
	g, _ := newAgentGame()
	g.Headless("play")

	obs := g.Step(Action{MouseX: 700, MouseY: 100, Click: true})
	obs = g.Step(Action{}) // scene switch applies on the next frame

	if obs.Scene != "menu" {
		t.Fatalf("clicking the button should switch scenes, got %q", obs.Scene)
	}
}

func TestStepIsDeterministic(t *testing.T) {
	run := func() Observation {
		g, _ := newAgentGame()
		g.Headless("play")
		var obs Observation
		for range 45 {
			obs = g.Step(Action{Keys: []Key{Right}})
		}
		return obs
	}
	a, b := run(), run()
	if len(a.Objects) != len(b.Objects) {
		t.Fatal("two identical runs should observe the same object count")
	}
	for i := range a.Objects {
		if a.Objects[i] != b.Objects[i] {
			t.Fatalf("run diverged at object %d: %+v vs %+v", i, a.Objects[i], b.Objects[i])
		}
	}
}

func TestAnyKeyByName(t *testing.T) {
	// Every keyboard key a player can press resolves by name, not just
	// the old 13-key list.
	for _, name := range []string{"J", "K", "Numpad1", "ShiftLeft", "Comma", "ArrowLeft", "Left", "Esc"} {
		if _, ok := keyNames[name]; !ok {
			t.Fatalf("key name %q should resolve", name)
		}
	}
}

func TestControlsResolveAndError(t *testing.T) {
	g, _ := newAgentGame()
	g.Controls(map[string]Key{"go-right": Right, "fire": Space})

	keys, err := g.resolveKeys([]any{"go-right", "fire", "J"})
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 3 || keys[0] != Right || keys[1] != Space {
		t.Fatalf("resolved = %v", keys)
	}

	// Unknown names error with the control list, instead of being
	// silently dropped.
	_, err = g.resolveKeys([]any{"warp"})
	if err == nil {
		t.Fatal("unknown key should error")
	}
	if want := "fire, go-right"; !containsStr(err.Error(), want) {
		t.Fatalf("error should list controls, got %q", err)
	}
}

func TestAgentStateInObservation(t *testing.T) {
	g, collected := newAgentGame()
	g.AgentState(func() any {
		return map[string]int{"collected": *collected}
	})
	g.Headless("play")

	obs := g.Observe()
	st, ok := obs.State.(map[string]int)
	if !ok || st["collected"] != 0 {
		t.Fatalf("state should be attached, got %#v", obs.State)
	}
	for range 120 {
		obs = g.Step(Action{Keys: []Key{Right}})
	}
	if st := obs.State.(map[string]int); st["collected"] != 1 {
		t.Fatalf("state should track the game, got %#v", obs.State)
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && strings.Contains(s, sub)
}

func TestDisallowAgentsBlocksHeadless(t *testing.T) {
	g, _ := newAgentGame()
	g.DisallowAgents()

	defer func() {
		if recover() == nil {
			t.Fatal("Headless should panic when agents are disallowed")
		}
	}()
	g.Headless("play")
}
