package collider

import (
	"math"
	"math/rand/v2"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// crowdScene builds the load a survivor game reaches mid-run: 1000
// live objects around a hero at the origin, with the camera on it.
//   - 400 enemies crowding toward the hero (respawned on a ring when
//     they arrive, so the crowd stays dense and steady),
//   - 150 projectiles flying out (wrapped back to the hero),
//   - 300 Visual gems lying around,
//   - 150 others: 130 tagged props and 20 solid walls (out beyond the
//     crowd, so no enemy piles up against one and the load stays
//     steady however many frames the benchmark runs).
func crowdScene() *Scene {
	g := New("bench", 800, 600)
	s := g.Scene("play")
	rng := rand.New(rand.NewPCG(1, 2))
	ring := func() (float64, float64) {
		a := rng.Float64() * 2 * math.Pi
		r := 200 + rng.Float64()*500
		return math.Cos(a) * r, math.Sin(a) * r
	}
	scatter := func() (float64, float64) {
		return (rng.Float64()*2 - 1) * 900, (rng.Float64()*2 - 1) * 900
	}

	hits := 0
	s.OnCollision("enemy", "shot", func(a, b *Object) { hits++ })
	for range 400 {
		e := s.Add(Rect(32, 40, Red).At(ring()).Tag("enemy").Layer(1))
		speed := 60 + rng.Float64()*60
		e.OnUpdate(func(dt float64) {
			e.MoveToward(0, 0, speed*dt)
			if e.X*e.X+e.Y*e.Y < 40*40 {
				e.At(ring())
			}
		})
	}
	for range 150 {
		a := rng.Float64() * 2 * math.Pi
		p := s.Add(Rect(12, 4, Yellow).At(0, 0).Tag("shot").Layer(2).Rotate(a))
		p.Vx, p.Vy = math.Cos(a)*400, math.Sin(a)*400
		p.OnUpdate(func(float64) {
			if p.X*p.X+p.Y*p.Y > 600*600 {
				p.At(0, 0)
			}
		})
	}
	for range 300 {
		s.Add(Rect(10, 10, Green).At(scatter()).Tag("gem").Visual())
	}
	for range 130 {
		s.Add(Rect(24, 24, Blue).At(scatter()).Tag("prop"))
	}
	for range 20 {
		a := rng.Float64() * 2 * math.Pi
		s.Add(Rect(64, 64, Blue).At(math.Cos(a)*850, math.Sin(a)*850).Solid())
	}
	s.Camera(0, 0)
	g.current = s
	s.activate()
	settle(s)
	return s
}

// settle runs the scene into its steady state after (re)starting.
func settle(s *Scene) {
	for range 60 {
		s.update(dt)
	}
}

// BenchmarkSceneUpdate measures one frame of the crowd scene. Every 600
// frames it rolls the scene back (untimed), so any b.N averages over
// the same ten-second window instead of a crowd that drifts.
func BenchmarkSceneUpdate(b *testing.B) {
	s := crowdScene()
	b.ReportAllocs()
	b.ResetTimer()
	for i := range b.N {
		if i > 0 && i%600 == 0 {
			b.StopTimer()
			s.restart()
			settle(s)
			b.StartTimer()
		}
		s.update(dt)
	}
}

func BenchmarkSceneDraw(b *testing.B) {
	s := crowdScene()
	screen := ebiten.NewImage(800, 600)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		s.draw(screen)
	}
}

func TestDrawCullingBounds(t *testing.T) {
	const w, h = 800.0, 600.0
	vx, vy := 1000.0, 1000.0 // camera looking at world (1400, 1300)
	for _, c := range []struct {
		name  string
		o     *Object
		ox    float64
		oy    float64
		shown bool
	}{
		{"inside", Rect(20, 20, Red).At(1400, 1300), vx, vy, true},
		{"far left", Rect(20, 20, Red).At(900, 1300), vx, vy, false},
		{"straddling the left edge", Rect(20, 20, Red).At(995, 1300), vx, vy, true},
		{"just past the right edge", Rect(20, 20, Red).At(1812, 1300), vx, vy, false},
		{"below", Rect(20, 20, Red).At(1400, 1700), vx, vy, false},
		// A long thin box turned 90 degrees reaches past its upright
		// box: the rotated bound keeps it.
		{"rotated into view", Rect(200, 4, Red).At(1400, 1690).Rotate(math.Pi / 2), vx, vy, true},
		{"rotated, still out", Rect(200, 4, Red).At(1400, 1720).Rotate(math.Pi / 2), vx, vy, false},
		// Fixed objects compare against the screen, not the world.
		{"fixed on screen", Rect(20, 20, Red).At(400, 300).Fixed(), 0, 0, true},
		{"fixed off screen", Rect(20, 20, Red).At(-50, 300).Fixed(), 0, 0, false},
		// Text and buttons are never culled.
		{"text far away", Text("HI").At(-5000, -5000), vx, vy, true},
		{"button far away", Button("OK").At(-5000, -5000), vx, vy, true},
	} {
		if got := c.o.onScreen(c.ox, c.oy, w, h); got != c.shown {
			t.Errorf("%s: onScreen = %v, want %v", c.name, got, c.shown)
		}
	}
}
