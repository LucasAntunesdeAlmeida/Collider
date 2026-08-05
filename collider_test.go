package collider

import "testing"

const dt = 1.0 / 60

func TestCollisionFiresOnEnterOnly(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")

	a := s.Add(Rect(10, 10, Red).At(0, 0))
	b := s.Add(Rect(10, 10, Red).At(100, 0))

	hits := 0
	a.OnCollision(func(*Object) { hits++ })

	s.update(dt)
	if hits != 0 {
		t.Fatalf("collision fired while objects are apart: hits = %d", hits)
	}

	b.At(5, 0)
	s.update(dt)
	s.update(dt)
	s.update(dt)
	if hits != 1 {
		t.Fatalf("enter event should fire exactly once during continuous overlap, got %d", hits)
	}

	b.At(100, 0)
	s.update(dt)
	b.At(5, 0)
	s.update(dt)
	if hits != 2 {
		t.Fatalf("re-entering should fire again, got %d", hits)
	}
}

func TestBothObjectsGetTheEvent(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")

	a := s.Add(Rect(10, 10, Red).At(0, 0))
	b := s.Add(Rect(10, 10, Red).At(5, 0))

	var aSaw, bSaw *Object
	a.OnCollision(func(o *Object) { aSaw = o })
	b.OnCollision(func(o *Object) { bSaw = o })

	s.update(dt)
	if aSaw != b || bSaw != a {
		t.Fatal("both objects should receive the collision event with the other as argument")
	}
}

func TestDestroyInsideCallbackIsSafe(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")

	a := s.Add(Rect(10, 10, Red).At(0, 0))
	b := s.Add(Rect(10, 10, Red).At(5, 0))

	a.OnCollision(func(other *Object) {
		other.Destroy()
		a.Destroy()
	})

	s.update(dt)
	s.update(dt)

	if len(s.objects) != 0 {
		t.Fatalf("destroyed objects should be removed at end of frame, %d left", len(s.objects))
	}
	_ = b
}

func TestVelocityAppliesEveryFrame(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")

	o := s.Add(Rect(10, 10, Red).At(0, 0))
	o.Vx = 60

	s.update(dt)
	if o.X <= 0 {
		t.Fatalf("velocity should move the object, X = %f", o.X)
	}
}

func TestObjectsAddedDuringUpdateJoinNextFrame(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")

	spawnerRan := 0
	spawner := s.Add(Rect(10, 10, Red).At(0, 0))
	spawner.OnUpdate(func(float64) {
		spawnerRan++
		if spawnerRan == 1 {
			s.Add(Rect(10, 10, Red).At(50, 50))
		}
	})

	s.update(dt)
	if len(s.objects) != 2 {
		t.Fatalf("spawned object should be in the scene, have %d objects", len(s.objects))
	}
}
