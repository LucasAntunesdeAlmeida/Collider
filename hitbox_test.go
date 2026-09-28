package collider

import (
	"math"
	"testing"
)

func TestHitboxDrivesCollisions(t *testing.T) {
	s := New("test", 800, 600).Scene("play")
	// Drawn 40x40 each, 30px apart: the drawings overlap, the 10x10
	// hitboxes do not.
	a := s.Add(Rect(40, 40, Red).At(100, 100).Hitbox(10, 10))
	b := s.Add(Rect(40, 40, Red).At(130, 100).Hitbox(10, 10))
	hits := 0
	a.OnCollision(func(*Object) { hits++ })
	s.update(dt)
	if hits != 0 {
		t.Fatal("overlapping drawings with apart hitboxes must not collide")
	}
	b.X = 108 // hitboxes now overlap by 2px
	s.update(dt)
	if hits != 1 {
		t.Fatalf("overlapping hitboxes should collide once, got %d", hits)
	}

	// Without Hitbox the drawn size is the collider, as before.
	c := s.Add(Rect(40, 40, Red).At(500, 100))
	s.Add(Rect(40, 40, Red).At(530, 100))
	plain := 0
	c.OnCollision(func(*Object) { plain++ })
	s.update(dt)
	if plain != 1 {
		t.Fatalf("objects without a hitbox collide with their drawn size, got %d", plain)
	}
}

func TestHitboxBiggerThanTheDrawing(t *testing.T) {
	s := New("test", 800, 600).Scene("play")
	// A 4px dot with a 100px pickup zone.
	zone := s.Add(Rect(4, 4, Red).At(100, 100).Hitbox(100, 100))
	s.Add(Rect(4, 4, Red).At(140, 100).Tag("gem"))
	hits := 0
	zone.OnCollisionWith("gem", func(*Object) { hits++ })
	s.update(dt)
	if hits != 1 {
		t.Fatalf("a hitbox larger than the drawing should reach the gem, got %d", hits)
	}
}

func TestHitboxDrivesSolids(t *testing.T) {
	s := New("test", 800, 600).Scene("play")
	s.Gravity(1000)
	s.Add(Rect(800, 40, Red).At(400, 580).Solid()) // top at 560
	// Drawn 20x60 (a tall sprite) with a centered 20x20 hitbox: the
	// hitbox rests on the floor, the drawing overhangs it by 20px.
	p := s.Add(Rect(20, 60, Red).At(400, 100).WithGravity().Hitbox(20, 20))
	for range 300 {
		s.update(dt)
	}
	if !p.Grounded() {
		t.Fatal("the hitbox should land on the floor")
	}
	if math.Abs(p.Y+10-560) > 0.5 {
		t.Fatalf("the hitbox bottom should rest on the floor top (560), got %f", p.Y+10)
	}

	// A solid's own hitbox is what pushes.
	s2 := New("test", 800, 600).Scene("play")
	s2.Add(Rect(200, 200, Red).At(400, 300).Solid().Hitbox(20, 20))
	o := s2.Add(Rect(10, 10, Red).At(440, 300)) // inside the drawing, outside the hitbox
	s2.update(dt)
	if o.X != 440 {
		t.Fatalf("an object outside a solid's hitbox must not be pushed, moved to %f", o.X)
	}
	o.X = 410 // overlaps the 20x20 hitbox by 5px
	s2.update(dt)
	if o.X != 415 {
		t.Fatalf("the solid's hitbox should push the object out to 415, got %f", o.X)
	}
}

func TestHitboxDrivesTouchingAndNear(t *testing.T) {
	s := New("test", 800, 600).Scene("play")
	hero := s.Add(Rect(42, 42, Red).At(0, 0).Hitbox(20, 20))
	// Drawn boxes overlap the hero's drawing, not its hitbox.
	s.Add(Rect(36, 36, Red).At(35, 0).Tag("e").Hitbox(20, 20))
	reached := s.Add(Rect(36, 36, Red).At(0, 18).Tag("e").Hitbox(20, 20))
	sameObjects(t, "Touching", hero.Touching("e"), []*Object{reached})

	// Near tests the hitbox: 10px wide at x=60 reaches 55, not 50.
	far := s.Add(Rect(80, 10, Red).At(60, 200).Tag("n").Hitbox(10, 10))
	if s.Near("n", 45, 200, 9) != nil {
		t.Fatal("Near should test the hitbox, not the drawing")
	}
	sameObjects(t, "Near", s.Near("n", 45, 200, 11), []*Object{far})
}

func TestHitboxChangeMidFrameReachesQueries(t *testing.T) {
	s := New("test", 800, 600).Scene("play")
	a := s.Add(Rect(10, 10, Red).At(0, 0))
	b := s.Add(Rect(10, 10, Red).At(200, 0).Tag("b"))
	if a.Touching("b") != nil { // builds the index for "b"
		t.Fatal("far apart")
	}
	b.Hitbox(500, 10) // now reaches across a, far past the query margin
	sameObjects(t, "Touching after Hitbox", a.Touching("b"), []*Object{b})
	sameObjects(t, "Near after Hitbox", s.Near("b", 0, 0, 1), []*Object{b})
}

func TestClicksHitTheDrawnBounds(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")
	card := s.Add(Rect(80, 80, Red).At(400, 300).Hitbox(10, 10))
	clicks := 0
	card.OnClick(func() { clicks++ })
	g.Headless("play")
	g.Step(Action{MouseX: 430, MouseY: 330, Click: true}) // on the drawing, off the hitbox
	if clicks != 1 {
		t.Fatalf("clicks should hit what the player sees, got %d", clicks)
	}
	g.Step(Action{MouseX: 450, MouseY: 300, Click: true}) // past the drawing
	if clicks != 1 {
		t.Fatal("a click past the drawing should miss")
	}
}

func TestObservationReportsTheHitbox(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")
	s.Add(Rect(42, 42, Red).At(100, 100).Tag("player").Hitbox(24, 30))
	s.Add(Rect(16, 16, Red).At(300, 100).Tag("gem"))
	g.Headless("play")
	for _, o := range g.Observe().Objects {
		switch o.Tag {
		case "player":
			if o.W != 24 || o.H != 30 {
				t.Fatalf("the player should report its hitbox 24x30, got %fx%f", o.W, o.H)
			}
		case "gem":
			if o.W != 16 || o.H != 16 {
				t.Fatalf("without a hitbox the drawn size is reported, got %fx%f", o.W, o.H)
			}
		}
	}
}

func TestRestartRestoresTheHitbox(t *testing.T) {
	s := New("test", 800, 600).Scene("play")
	a := s.Add(Rect(40, 40, Red).Hitbox(10, 12))
	b := s.Add(Rect(40, 40, Red))
	s.activate()

	a.Hitbox(30, 30)
	b.Hitbox(5, 5)
	s.restart()

	if box := a.box(); box.W != 10 || box.H != 12 {
		t.Fatalf("restart should restore the setup hitbox, got %fx%f", box.W, box.H)
	}
	if box := b.box(); box.W != 40 || box.H != 40 {
		t.Fatalf("restart should drop a hitbox set during play, got %fx%f", box.W, box.H)
	}
}

func TestDrawingUsesTheDrawnSize(t *testing.T) {
	o := Rect(100, 50, Red).At(0, 0).Hitbox(2, 2)
	m := o.geoM(0, 0, 1, 1)
	if x, y := m.Apply(1, 1); !near(x, 50) || !near(y, 25) {
		t.Fatalf("the drawing should keep its 100x50 size, corner at (%f, %f)", x, y)
	}
	// Culling uses the drawing too: the hitbox is far off screen, the
	// drawing reaches into it.
	o.At(-45, 300)
	if !o.onScreen(0, 0, 800, 600) {
		t.Fatal("an object whose drawing reaches the screen must be drawn")
	}
}
