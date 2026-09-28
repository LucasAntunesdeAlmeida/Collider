package collider

import (
	"slices"
	"testing"
)

func TestLayersDrawBottomToTop(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")

	hud := s.Add(Text("HUD").Layer(10))
	a := s.Add(Rect(10, 10, Red))
	ground := s.Add(Rect(10, 10, Red).Layer(-1))
	b := s.Add(Rect(10, 10, Red))
	fx := s.Add(Rect(10, 10, Red).Layer(3))

	want := []*Object{ground, a, b, fx, hud}
	if got := s.drawOrder(); !slices.Equal(got, want) {
		t.Fatal("draw order should be by layer, then by the order objects were added")
	}
}

func TestDrawOrderSkipsDestroyedObjects(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")
	a := s.Add(Rect(10, 10, Red))
	b := s.Add(Rect(10, 10, Red))
	a.Destroy()

	if got := s.drawOrder(); !slices.Equal(got, []*Object{b}) {
		t.Fatalf("destroyed objects must not be drawn, got %d objects", len(got))
	}
}

func TestLayerChangesApplyAtRuntime(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")
	a := s.Add(Rect(10, 10, Red))
	b := s.Add(Rect(10, 10, Red))

	a.Layer(1) // bring to front mid-game
	if got := s.drawOrder(); !slices.Equal(got, []*Object{b, a}) {
		t.Fatal("changing a layer at runtime should reorder drawing")
	}
}

func TestClicksHitTheTopmostLayer(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")

	top := s.Add(Button("FRONT").At(400, 300).Layer(5)) // added first, drawn last
	under := s.Add(Button("BACK").At(400, 300))
	tops, unders := 0, 0
	top.OnClick(func() { tops++ })
	under.OnClick(func() { unders++ })

	g.Headless("play")
	g.Step(Action{MouseX: 400, MouseY: 300, Click: true})
	if tops != 1 || unders != 0 {
		t.Fatalf("the click should go to the object drawn on top, got top=%d under=%d", tops, unders)
	}
}

func TestRestartRestoresLayers(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")
	o := s.Add(Rect(10, 10, Red).Layer(2))
	s.activate()

	o.Layer(9)
	s.restart()
	if o.layer != 2 {
		t.Fatalf("restart should restore the setup layer, got %d", o.layer)
	}
}
