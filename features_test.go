package collider

import (
	"bytes"
	"image"
	"image/png"
	"math"
	"testing"
	"testing/fstest"
)

func TestGravityLandingAndGrounded(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")
	s.Gravity(1000)

	s.Add(Rect(800, 40, Red).At(400, 580).Solid())
	p := s.Add(Rect(20, 20, Red).At(400, 100).WithGravity())

	for range 300 {
		s.update(dt)
	}

	if !p.Grounded() {
		t.Fatal("object should be grounded after falling onto a solid floor")
	}
	if p.Vy != 0 {
		t.Fatalf("vertical velocity should be zeroed on landing, got %f", p.Vy)
	}
	floorTop := 580.0 - 20
	if math.Abs((p.Y+10)-floorTop) > 0.5 {
		t.Fatalf("object should rest on the floor top (%f), bottom is at %f", floorTop, p.Y+10)
	}
}

func TestJumpClearsGrounded(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")
	s.Gravity(1000)

	s.Add(Rect(800, 40, Red).At(400, 580).Solid())
	p := s.Add(Rect(20, 20, Red).At(400, 540).WithGravity())

	for range 60 {
		s.update(dt)
	}
	if !p.Grounded() {
		t.Fatal("object should be grounded before jumping")
	}

	p.Vy = -500
	s.update(dt)
	s.update(dt)
	if p.Grounded() {
		t.Fatal("object should not be grounded mid-jump")
	}
}

func TestOnCollisionWithFiltersByTag(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")

	p := s.Add(Rect(10, 10, Red).At(0, 0))
	s.Add(Rect(10, 10, Red).At(5, 0).Tag("enemy"))
	s.Add(Rect(10, 10, Red).At(-5, 0).Tag("coin"))

	enemies, coins, all := 0, 0, 0
	p.OnCollisionWith("enemy", func(*Object) { enemies++ })
	p.OnCollisionWith("coin", func(*Object) { coins++ })
	p.OnCollision(func(*Object) { all++ })

	s.update(dt)

	if enemies != 1 || coins != 1 {
		t.Fatalf("tag callbacks should fire once each, got enemies=%d coins=%d", enemies, coins)
	}
	if all != 2 {
		t.Fatalf("generic callback should fire for both, got %d", all)
	}
}

func TestSceneRuleOrdersArgsByTag(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")

	b := s.Add(Rect(10, 10, Red).At(5, 0).Tag("bullet"))
	z := s.Add(Rect(10, 10, Red).At(0, 0).Tag("zombie"))

	var gotZ, gotB *Object
	s.OnCollision("zombie", "bullet", func(a, o *Object) { gotZ, gotB = a, o })

	s.update(dt)

	if gotZ != z || gotB != b {
		t.Fatal("rule callback should receive objects in tag order regardless of add order")
	}
}

func TestCountTag(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")

	for i := range 3 {
		s.Add(Rect(10, 10, Red).At(float64(i)*100, 0).Tag("brick"))
	}
	s.Add(Rect(10, 10, Red).At(500, 0))

	if got := s.Count("brick"); got != 3 {
		t.Fatalf("Count = %d, want 3", got)
	}
	s.objects[0].Destroy()
	s.update(dt)
	if got := s.Count("brick"); got != 2 {
		t.Fatalf("Count after destroy = %d, want 2", got)
	}
}

func TestLifeTimeDestroys(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")

	o := s.Add(Rect(10, 10, Red).At(0, 0))
	o.LifeTime(0.05)

	for range 10 {
		s.update(dt)
	}
	if len(s.objects) != 0 {
		t.Fatal("object should self-destroy after its lifetime")
	}
}

func TestEveryRepeatsAndAfterFiresOnce(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")

	every, after := 0, 0
	s.Every(0.1, func() { every++ })
	s.After(0.25, func() { after++ })

	for range 60 {
		s.update(dt)
	}

	if every < 9 || every > 10 {
		t.Fatalf("Every(0.1) over 1s should fire 9 or 10 times, got %d", every)
	}
	if after != 1 {
		t.Fatalf("After should fire exactly once, got %d", after)
	}
}

func TestVelocityTowardNormalizes(t *testing.T) {
	o := &Object{X: 0, Y: 0}
	o.VelocityToward(3, 4, 100)
	if math.Abs(o.Vx-60) > 1e-9 || math.Abs(o.Vy-80) > 1e-9 {
		t.Fatalf("velocity should be (60, 80), got (%f, %f)", o.Vx, o.Vy)
	}
}

func TestMoveTowardStopsAtTarget(t *testing.T) {
	o := &Object{X: 0, Y: 0}
	o.MoveToward(10, 0, 25)
	if o.X != 10 || o.Y != 0 {
		t.Fatalf("object should stop exactly on the target, got (%f, %f)", o.X, o.Y)
	}
}

func TestRestartRestoresInitialState(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")

	o := s.Add(Rect(10, 10, Red).At(50, 50))
	s.Every(1, func() {})
	s.activate()

	o.X = 999
	o.Destroy()
	s.Add(Rect(10, 10, Red).At(1, 1)) // runtime spawn
	s.After(1, func() {})             // runtime timer
	s.update(dt)

	s.restart()

	if len(s.objects) != 1 || s.objects[0] != o {
		t.Fatalf("restart should keep exactly the setup objects, got %d", len(s.objects))
	}
	if o.dead || o.X != 50 {
		t.Fatalf("restart should revive and reposition setup objects, dead=%v X=%f", o.dead, o.X)
	}
	if len(s.timers) != 1 || !s.timers[0].initial {
		t.Fatalf("restart should keep only setup timers, got %d", len(s.timers))
	}
}

func TestTextObjectsDoNotCollide(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")

	s.Add(Text("score").At(0, 0))
	o := s.Add(Rect(50, 50, Red).At(0, 0))

	hits := 0
	o.OnCollision(func(*Object) { hits++ })

	s.update(dt)

	if hits != 0 {
		t.Fatal("text objects must never take part in collisions")
	}
}

func TestUseAssetsLoadsFromEmbeddedFS(t *testing.T) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 3, 5))); err != nil {
		t.Fatal(err)
	}
	UseAssets(fstest.MapFS{
		"sprites/p.png": &fstest.MapFile{Data: buf.Bytes()},
	})
	defer UseAssets(nil)

	g := New("test", 800, 600)
	s := g.Scene("play")
	o := s.Add(Sprite("sprites/p.png"))

	if o.w != 3 || o.h != 5 {
		t.Fatalf("sprite should load from the embedded FS with its real size, got %fx%f", o.w, o.h)
	}
}

func TestAnimationFrameTiming(t *testing.T) {
	o := &Object{}
	o.anims = map[string]*animation{
		"run":   {frames: 4, fps: 10},
		"death": {frames: 4, fps: 10},
	}

	o.play("run", false)
	o.animate(0.25)
	if f := o.frameIndex(o.anims["run"]); f != 2 {
		t.Fatalf("0.25s at 10fps should be frame 2, got %d", f)
	}
	o.animate(0.2)
	if f := o.frameIndex(o.anims["run"]); f != 0 {
		t.Fatalf("looping animation should wrap to frame 0, got %d", f)
	}

	elapsed := o.animTime
	o.play("run", false)
	if o.animTime != elapsed {
		t.Fatal("re-playing the active animation must not reset it")
	}

	o.play("death", true)
	o.animate(2)
	if f := o.frameIndex(o.anims["death"]); f != 3 {
		t.Fatalf("PlayOnce should hold the last frame (3), got %d", f)
	}
}

func TestVisualObjectsDoNotCollide(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")

	s.Add(Rect(200, 200, Blue).At(0, 0).Visual()) // scenery
	o := s.Add(Rect(20, 20, Red).At(0, 0))

	hits := 0
	o.OnCollision(func(*Object) { hits++ })
	s.update(dt)

	if hits != 0 {
		t.Fatal("Visual objects must never take part in collisions")
	}
}

func TestButtonSizesToItsLabelAndClicks(t *testing.T) {
	g := New("test", 800, 600)
	s := g.Scene("play")

	b := s.Add(Button("PLAY").At(400, 300))
	if b.w <= buttonPadX*2 || b.h <= buttonPadY*2 {
		t.Fatalf("button should size around its label, got %fx%f", b.w, b.h)
	}

	clicked := 0
	b.OnClick(func() { clicked++ })
	if !b.contains(400, 300) {
		t.Fatal("button bounds should contain its own center")
	}
	if b.contains(400-b.w, 300) {
		t.Fatal("button bounds should not extend a full width to the left")
	}

	// Clicks reach buttons through the normal object path.
	g.Headless("play")
	g.Step(Action{MouseX: 400, MouseY: 300, Click: true})
	if clicked != 1 {
		t.Fatalf("button should fire OnClick once, got %d", clicked)
	}
}

func TestShuffleKeepsElements(t *testing.T) {
	in := []int{1, 2, 3, 4, 5, 6, 7, 8}
	out := Shuffle(in)
	if len(out) != len(in) {
		t.Fatalf("length changed: %d", len(out))
	}
	counts := map[int]int{}
	for _, v := range out {
		counts[v]++
	}
	for _, v := range in {
		if counts[v] != 1 {
			t.Fatalf("element %d lost or duplicated", v)
		}
	}
}
