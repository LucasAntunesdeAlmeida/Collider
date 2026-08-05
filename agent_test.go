package collider

import (
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
