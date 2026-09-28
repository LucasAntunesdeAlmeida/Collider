package collider

import (
	"cmp"
	"math"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/LucasAntunesdeAlmeida/collider/internal/physics"
)

// bruteNear is the reference Near: every object checked, stable sort.
func bruteNear(s *Scene, tag string, x, y, r float64) []*Object {
	var out []*Object
	for _, o := range s.objects {
		if o.dead || o.fixed || o.tag != tag {
			continue
		}
		box := o.box()
		dx := max(math.Abs(o.X-x)-box.W/2, 0)
		dy := max(math.Abs(o.Y-y)-box.H/2, 0)
		if dx*dx+dy*dy <= r*r {
			out = append(out, o)
		}
	}
	slices.SortStableFunc(out, func(a, b *Object) int {
		return cmp.Compare(math.Hypot(a.X-x, a.Y-y), math.Hypot(b.X-x, b.Y-y))
	})
	return out
}

// bruteTouching is the reference Touching.
func bruteTouching(o *Object, tag string) []*Object {
	if o.visual || o.dead {
		return nil
	}
	var out []*Object
	for _, other := range o.scene.objects {
		if other != o && !other.dead && !other.visual && other.tag == tag &&
			physics.Overlaps(o.box(), other.box()) {
			out = append(out, other)
		}
	}
	return out
}

// randomScene fills a scene with a seeded mix of tags, sizes (a few
// huge), hitboxes apart from the drawn size (some bigger), visual,
// fixed and destroyed objects, some moving.
func randomScene(seed uint64, n int) (*Scene, *rand.Rand) {
	rng := rand.New(rand.NewPCG(seed, seed))
	s := New("test", 800, 600).Scene("play")
	tags := []string{"a", "b", ""}
	for range n {
		w, h := 4+rng.Float64()*60, 4+rng.Float64()*60
		if rng.IntN(20) == 0 {
			w, h = 300+rng.Float64()*600, 10+rng.Float64()*40
		}
		o := s.Add(Rect(w, h, Red).At(rng.Float64()*1600-800, rng.Float64()*1600-800).Tag(tags[rng.IntN(3)]))
		o.Vx, o.Vy = rng.Float64()*400-200, rng.Float64()*400-200
		if rng.IntN(4) == 0 {
			o.Hitbox(w*(0.2+rng.Float64()*1.6), h*(0.2+rng.Float64()*1.6))
		}
		switch rng.IntN(10) {
		case 0:
			o.Visual()
		case 1:
			o.Fixed()
		case 2:
			o.Destroy()
		}
	}
	return s, rng
}

func sameObjects(t *testing.T, what string, got, want []*Object) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Fatalf("%s: got %d objects, want %d (or a different order)", what, len(got), len(want))
	}
}

func TestNearAndTouchingMatchBruteForce(t *testing.T) {
	for seed := range uint64(8) {
		s, rng := randomScene(seed, 300)
		for frame := range 4 {
			for range 40 {
				x, y, r := rng.Float64()*2000-1000, rng.Float64()*2000-1000, rng.Float64()*400
				for _, tag := range []string{"a", "b", "none"} {
					sameObjects(t, "Near", s.Near(tag, x, y, r), bruteNear(s, tag, x, y, r))
				}
			}
			for _, o := range s.objects {
				sameObjects(t, "Touching", o.Touching("a"), bruteTouching(o, "a"))
			}
			if frame < 3 {
				s.update(dt * 3) // objects move; destroyed ones are flushed
			}
		}
	}
}

func TestNearSortsByCenterDistanceTiesInSceneOrder(t *testing.T) {
	s := New("test", 800, 600).Scene("play")
	far := s.Add(Rect(10, 10, Red).At(90, 0).Tag("e"))
	tieA := s.Add(Rect(10, 10, Red).At(0, 50).Tag("e"))
	nearest := s.Add(Rect(10, 10, Red).At(10, 0).Tag("e"))
	tieB := s.Add(Rect(10, 10, Red).At(-50, 0).Tag("e"))
	got := s.Near("e", 0, 0, 100)
	sameObjects(t, "Near order", got, []*Object{nearest, tieA, tieB, far})
}

func TestNearTestsTheBoxNotTheCenter(t *testing.T) {
	s := New("test", 800, 600).Scene("play")
	// A long wall whose center is 500px away but whose end is 5px away.
	wall := s.Add(Rect(1000, 20, Red).At(500, 0).Tag("wall"))
	sameObjects(t, "big box", s.Near("wall", -5, 0, 10), []*Object{wall})
	if s.Near("wall", -20, 0, 10) != nil {
		t.Fatal("a box out of the circle should not be found")
	}
	// Corner: the closest point is the corner, not an edge midpoint.
	box := s.Add(Rect(20, 20, Red).At(0, 100).Tag("box"))
	if s.Near("box", 20, 120, 14) != nil { // corner (10, 110) is 14.14 away
		t.Fatal("the circle misses the corner")
	}
	sameObjects(t, "corner", s.Near("box", 20, 120, 14.2), []*Object{box})
}

func TestNearExclusions(t *testing.T) {
	s := New("test", 800, 600).Scene("play")
	gem := s.Add(Rect(10, 10, Red).At(0, 0).Tag("gem").Visual())
	s.Add(Rect(10, 10, Red).At(0, 0).Tag("gem").Fixed())
	s.Add(Rect(10, 10, Red).At(0, 0).Tag("other"))
	s.Add(Rect(10, 10, Red).At(0, 0).Tag("gem")).Destroy()
	sameObjects(t, "Near", s.Near("gem", 0, 0, 50), []*Object{gem})
}

func TestTouchingExclusions(t *testing.T) {
	s := New("test", 800, 600).Scene("play")
	p := s.Add(Rect(20, 20, Red).At(0, 0).Tag("e"))
	hit := s.Add(Rect(20, 20, Red).At(10, 10).Tag("e"))
	s.Add(Rect(20, 20, Red).At(5, 5).Tag("e").Visual())
	s.Add(Rect(20, 20, Red).At(5, 5).Tag("e").Fixed())
	s.Add(Text("E").At(0, 0).Tag("e"))
	s.Add(Rect(20, 20, Red).At(5, 5).Tag("f"))
	s.Add(Rect(20, 20, Red).At(20, 0).Tag("e")) // edges touch: no overlap
	s.Add(Rect(20, 20, Red).At(0, 0).Tag("e")).Destroy()
	sameObjects(t, "Touching", p.Touching("e"), []*Object{hit})
	ghost := s.Add(Rect(20, 20, Red).At(0, 0).Tag("e").Visual())
	if ghost.Touching("e") != nil {
		t.Fatal("a visual object never touches anything")
	}
	if Rect(20, 20, Red).Touching("e") != nil {
		t.Fatal("an object outside any scene touches nothing")
	}
}

func TestQueriesSeeObjectsAddedMovedAndRetaggedThisFrame(t *testing.T) {
	s := New("test", 800, 600).Scene("play")
	a := s.Add(Rect(10, 10, Red).At(0, 0).Tag("e"))
	sameObjects(t, "built", s.Near("e", 0, 0, 20), []*Object{a}) // index built now

	b := s.Add(Rect(10, 10, Red).At(5, 0).Tag("e")) // added after the build
	sameObjects(t, "added", s.Near("e", 0, 0, 20), []*Object{a, b})

	a.X = 30 // a small move since the build: found where it is now
	sameObjects(t, "moved", s.Near("e", 0, 0, 20), []*Object{b})
	sameObjects(t, "moved", s.Near("e", 30, 0, 1), []*Object{a})

	b.Tag("f")
	if s.Near("e", 0, 0, 20) != nil {
		t.Fatal("a re-tagged object leaves its old tag at once")
	}
	sameObjects(t, "retagged", s.Near("f", 0, 0, 20), []*Object{b})
}

func TestQueryIndexRebuildsAcrossUpdates(t *testing.T) {
	s := New("test", 800, 600).Scene("play")
	o := s.Add(Rect(10, 10, Red).At(0, 0).Tag("e"))
	o.Vx = 30000 // 500px in one frame, far beyond the lookup margin
	s.Near("e", 0, 0, 10)
	s.update(dt)
	if s.Near("e", 0, 0, 10) != nil {
		t.Fatal("the old position should be empty after the update")
	}
	sameObjects(t, "after update", s.Near("e", 500, 0, 10), []*Object{o})

	// A query inside a collision callback sees post-motion positions,
	// even when an update callback built the index before the motion.
	o.OnUpdate(func(float64) { s.Near("e", 0, 0, 1) })
	var seen []*Object
	s.Add(Rect(10, 10, Red).At(1000, 0).Tag("wall"))
	s.OnCollision("e", "wall", func(e, _ *Object) { seen = s.Near("e", e.X, e.Y, 1) })
	s.update(dt)
	sameObjects(t, "in callback", seen, []*Object{o})

	o.Destroy()
	s.update(dt) // flushed: indices shift
	if s.Near("e", o.X, o.Y, 50) != nil {
		t.Fatal("a flushed object is gone from the index")
	}
}

func benchScene() (*Scene, *Object) {
	rng := rand.New(rand.NewPCG(1, 2))
	s := New("test", 800, 600).Scene("play")
	for range 400 {
		s.Add(Rect(24, 24, Red).At(rng.Float64()*2000-1000, rng.Float64()*2000-1000).Tag("enemy"))
	}
	p := s.Add(Rect(40, 40, Red).At(0, 0).Tag("player"))
	return s, p
}

func BenchmarkNear(b *testing.B) {
	s, p := benchScene()
	for b.Loop() {
		s.Near("enemy", p.X, p.Y, 300)
	}
}

func BenchmarkTouching(b *testing.B) {
	_, p := benchScene()
	for b.Loop() {
		p.Touching("enemy")
	}
}

// BenchmarkNearFreshIndex pays for one index build per query: the cost
// of the first query of a frame.
func BenchmarkNearFreshIndex(b *testing.B) {
	s, p := benchScene()
	for b.Loop() {
		s.index.invalidate()
		s.Near("enemy", p.X, p.Y, 300)
	}
}
