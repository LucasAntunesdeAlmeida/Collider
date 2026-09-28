package collider

import "testing"

func TestCameraDefaultKeepsScreenCoordinates(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")

	if x, y := s.view(); x != 0 || y != 0 {
		t.Fatalf("a scene that never moves its camera must draw world as screen, view = (%f, %f)", x, y)
	}
}

func TestCameraCentersTheViewOnAWorldPoint(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")
	o := s.Add(Rect(10, 10, Red).At(3000, -500))

	s.Camera(3000, -500)
	vx, vy := s.origin(o)
	if left, top := o.X-o.w/2-vx, o.Y-o.h/2-vy; left != 395 || top != 295 {
		t.Fatalf("the object under the camera should draw at the screen center, got top-left (%f, %f)", left, top)
	}

	s.Camera(100.4, 200.6)
	if x, y := s.view(); x != -300 || y != -99 {
		t.Fatalf("the view origin should round to whole pixels, got (%f, %f)", x, y)
	}
}

func TestFixedObjectsIgnoreTheCamera(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")
	hud := s.Add(Text("HP").At(60, 20).Fixed())

	s.Camera(5000, 5000)
	if x, y := s.origin(hud); x != 0 || y != 0 {
		t.Fatalf("fixed objects draw in screen space, origin = (%f, %f)", x, y)
	}
}

func TestFixedObjectsDoNotCollide(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")

	s.Add(Rect(100, 100, Blue).At(0, 0).Fixed())
	o := s.Add(Rect(20, 20, Red).At(0, 0))

	hits := 0
	o.OnCollision(func(*Object) { hits++ })
	s.update(dt)

	if hits != 0 {
		t.Fatal("fixed objects live on the screen, not in the world, and must never collide")
	}
}

func TestClicksHitWorldObjectsUnderTheCamera(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")

	world := s.Add(Rect(40, 40, Red).At(2000, 1000))
	button := s.Add(Button("PAUSE").At(400, 300).Fixed())
	var sceneX, sceneY float64
	s.OnClick(func(x, y float64) { sceneX, sceneY = x, y })

	worldHits, buttonHits := 0, 0
	world.OnClick(func() { worldHits++ })
	button.OnClick(func() { buttonHits++ })

	s.Camera(2000, 1100) // the world object sits 100px above center
	g.Headless("play")

	g.Step(Action{MouseX: 400, MouseY: 200, Click: true})
	if worldHits != 1 {
		t.Fatalf("a click where a world object appears should hit it, got %d", worldHits)
	}
	if sceneX != 2000 || sceneY != 1000 {
		t.Fatalf("scene OnClick should receive world coordinates, got (%f, %f)", sceneX, sceneY)
	}

	g.Step(Action{MouseX: 400, MouseY: 300, Click: true})
	if buttonHits != 1 {
		t.Fatalf("a fixed button should be hit at its screen position, got %d", buttonHits)
	}
	if worldHits != 1 {
		t.Fatal("the world object is no longer under the cursor")
	}
}

func TestRestartRestoresTheCamera(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")
	s.Camera(50, 60)
	s.activate()

	s.Camera(9000, 9000)
	s.restart()

	if s.camX != 50 || s.camY != 60 {
		t.Fatalf("restart should put the camera back where setup left it, got (%f, %f)", s.camX, s.camY)
	}
}

func TestObservationMarksFixedObjects(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")
	s.Add(Text("SCORE").At(60, 20).Fixed())
	s.Add(Rect(10, 10, Red).At(1, 1).Tag("player"))
	g.Headless("play")

	var fixed, world int
	for _, o := range g.Observe().Objects {
		if o.Fixed {
			fixed++
		} else {
			world++
		}
	}
	if fixed != 1 || world != 1 {
		t.Fatalf("observation should flag the HUD text as fixed, got fixed=%d world=%d", fixed, world)
	}
}
