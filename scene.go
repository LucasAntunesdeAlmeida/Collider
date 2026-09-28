package collider

import (
	"math"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/LucasAntunesdeAlmeida/collider/internal/physics"
)

// Scene is a screen of the game: a level, a menu, a game-over screen.
// It holds objects and runs their events every frame.
type Scene struct {
	game    *Game
	name    string
	objects []*Object

	// activated flips on first entry; at that moment the scene snapshots
	// its objects so Restart can roll back to them.
	activated bool
	initial   []*Object

	gravity   float64
	musicPath string

	// camX, camY is the world point at the center of the screen. The
	// default is the screen center itself, so a scene that never moves
	// its camera draws world coordinates as screen coordinates.
	camX, camY         float64
	initCamX, initCamY float64

	timers   []*timer
	onClick  []func(x, y float64)
	onUpdate []func(dt float64)
	rules    []collisionRule

	// touching tracks currently overlapping pairs so collision events
	// fire on enter, not on every frame of overlap. touchNext is the
	// map being filled this frame; the two swap, so no frame allocates
	// a fresh one.
	touching  map[pair]struct{}
	touchNext map[pair]struct{}

	// Per-frame scratch for collide and resolveSolids, reused so a
	// steady scene allocates nothing: the broad-phase grid, the scene
	// index of each collidable box, the boxes, and the solid objects.
	grid     physics.Grid
	collIdx  []int
	collBox  []physics.Box
	solidBuf []*Object

	// order is the draw list, rebuilt every frame and reused so drawing
	// allocates nothing.
	order []*Object

	// index serves Near and Touching (see query.go).
	index queryIndex
}

type pair [2]*Object

type collisionRule struct {
	tagA, tagB string
	fn         func(a, b *Object)
}

type timer struct {
	period    float64
	remaining float64
	repeat    bool
	// initial timers were registered during setup (before the scene
	// first ran); they survive Restart, runtime ones are dropped.
	initial bool
	done    bool
	fn      func()
}

func newScene(g *Game, name string) *Scene {
	return &Scene{
		game:      g,
		name:      name,
		camX:      g.Width() / 2,
		camY:      g.Height() / 2,
		touching:  map[pair]struct{}{},
		touchNext: map[pair]struct{}{},
	}
}

// Camera centers the view on a world point: every object draws shifted
// so that (x, y) lands in the middle of the screen, and clicks land in
// the world where they appear to. Call it every frame to follow the
// player. Objects marked Fixed ignore the camera. The default camera
// looks at the screen center, where world and screen coordinates are
// the same.
func (s *Scene) Camera(x, y float64) {
	s.camX, s.camY = x, y
}

// view returns the world position of the screen's top-left corner.
// It is rounded to whole pixels so pixel art never shimmers while the
// camera glides.
func (s *Scene) view() (x, y float64) {
	return math.Round(s.camX - s.game.Width()/2), math.Round(s.camY - s.game.Height()/2)
}

// origin returns the view origin an object draws and is clicked
// against: the camera's for world objects, the screen's for Fixed ones.
func (s *Scene) origin(o *Object) (x, y float64) {
	if o.fixed {
		return 0, 0
	}
	return s.view()
}

// Add puts an object into the scene. It is safe to call at any time,
// including from inside callbacks; the object joins this frame.
func (s *Scene) Add(o *Object) *Object {
	o.scene = s
	o.resolve(s.game)
	s.objects = append(s.objects, o)
	return o
}

// Music sets looping background music (wav or ogg) that plays while the
// scene is active.
func (s *Scene) Music(path string) {
	s.musicPath = path
}

// Gravity sets downward acceleration in pixels per second squared for
// objects marked WithGravity.
func (s *Scene) Gravity(g float64) {
	s.gravity = g
}

// Every runs fn repeatedly, every sec seconds, while the scene is active.
func (s *Scene) Every(sec float64, fn func()) {
	s.timers = append(s.timers, &timer{
		period: sec, remaining: sec, repeat: true, initial: !s.activated, fn: fn,
	})
}

// After runs fn once, sec seconds from now, while the scene is active.
func (s *Scene) After(sec float64, fn func()) {
	s.timers = append(s.timers, &timer{
		period: sec, remaining: sec, initial: !s.activated, fn: fn,
	})
}

// OnClick fires on every click anywhere in the scene, before any
// object-level click handling. The position is in world coordinates,
// so under a moving camera it is where the click lands in the world.
func (s *Scene) OnClick(fn func(x, y float64)) {
	s.onClick = append(s.onClick, fn)
}

// OnUpdate runs every frame while the scene is active, before object
// updates. The place for scene-wide input like a fullscreen toggle.
func (s *Scene) OnUpdate(fn func(dt float64)) {
	s.onUpdate = append(s.onUpdate, fn)
}

// OnCollision declares a rule for every current and future pair of
// objects with these tags. The callback receives the objects in tag
// order: the tagA object first.
func (s *Scene) OnCollision(tagA, tagB string, fn func(a, b *Object)) {
	s.rules = append(s.rules, collisionRule{tagA: tagA, tagB: tagB, fn: fn})
}

// Count reports how many objects with this tag are alive.
func (s *Scene) Count(tag string) int {
	n := 0
	for _, o := range s.objects {
		if !o.dead && o.tag == tag {
			n++
		}
	}
	return n
}

// activate snapshots the scene the first time it becomes current.
func (s *Scene) activate() {
	if s.activated {
		return
	}
	s.activated = true
	s.initCamX, s.initCamY = s.camX, s.camY
	s.initial = slices.Clone(s.objects)
	for _, o := range s.initial {
		o.saveState()
	}
}

// restart rolls the scene back to its initial snapshot: setup objects
// restored (including destroyed ones), runtime spawns dropped, initial
// timers re-armed, runtime timers dropped, camera back where it was.
func (s *Scene) restart() {
	if !s.activated {
		return
	}
	s.camX, s.camY = s.initCamX, s.initCamY
	s.objects = slices.Clone(s.initial)
	for _, o := range s.objects {
		o.restoreState()
	}
	s.touching = map[pair]struct{}{}
	s.index.invalidate()
	kept := s.timers[:0]
	for _, t := range s.timers {
		if t.initial {
			t.remaining = t.period
			t.done = false
			kept = append(kept, t)
		}
	}
	s.timers = kept
}

// update runs one frame: input, timers, per-object logic, motion,
// collisions, solid resolution, cleanup.
func (s *Scene) update(dt float64) {
	s.index.invalidate()
	s.clicks()
	s.tick(dt)

	for _, fn := range slices.Clone(s.onUpdate) {
		fn(dt)
	}

	// Snapshot the length so objects added during callbacks start next frame.
	n := len(s.objects)
	for i := range n {
		o := s.objects[i]
		if o.dead {
			continue
		}
		for _, fn := range o.onUpdate {
			fn(dt)
		}
	}

	for _, o := range s.objects {
		if o.dead {
			continue
		}
		o.animate(dt)
		if o.gravity {
			o.Vy += s.gravity * dt
		}
		o.X += o.Vx * dt
		o.Y += o.Vy * dt
		if o.hasLife {
			o.life -= dt
			if o.life <= 0 {
				o.dead = true
			}
		}
	}
	s.index.invalidate() // everything moved

	s.collide()
	s.resolveSolids()
	s.flush()
}

// tick advances timers. Timers added from inside a callback start
// counting next frame.
func (s *Scene) tick(dt float64) {
	n := len(s.timers)
	for i := range n {
		t := s.timers[i]
		if t.done {
			continue
		}
		t.remaining -= dt
		if t.remaining > 0 {
			continue
		}
		if t.repeat {
			t.remaining += t.period
		} else {
			t.done = true
		}
		t.fn()
	}
	// Fired one-shots are removed, unless they must survive for Restart.
	s.timers = slices.DeleteFunc(s.timers, func(t *timer) bool {
		return t.done && !t.initial
	})
}

// clicks fires scene handlers on any click, then OnClick on the topmost
// object under the cursor. Fixed objects are hit in screen space, the
// rest in the world under the camera.
func (s *Scene) clicks() {
	if !s.game.input.clickJustPressed() {
		return
	}
	x, y := s.game.Mouse()
	vx, vy := s.view()
	for _, fn := range slices.Clone(s.onClick) {
		fn(x+vx, y+vy)
	}
	for _, o := range slices.Backward(s.drawOrder()) {
		if len(o.onClick) == 0 {
			continue
		}
		ox, oy := s.origin(o)
		if o.contains(x+ox, y+oy) {
			for _, fn := range o.onClick {
				fn()
			}
			return
		}
	}
}

// collide detects overlaps (spatial hash broad phase, exact AABB narrow
// phase) and fires enter events: object callbacks and scene tag rules.
func (s *Scene) collide() {
	s.collIdx = s.collIdx[:0]
	s.collBox = s.collBox[:0]
	for i, o := range s.objects {
		if o.dead || o.visual {
			continue
		}
		s.collIdx = append(s.collIdx, i)
		s.collBox = append(s.collBox, o.box())
	}

	now := s.touchNext
	clear(now)
	for _, c := range s.grid.Pairs(s.collBox) {
		a := s.objects[s.collIdx[c[0]]]
		b := s.objects[s.collIdx[c[1]]]
		if a.dead || b.dead {
			continue
		}
		if !physics.Overlaps(a.box(), b.box()) {
			continue
		}
		p := pair{a, b}
		now[p] = struct{}{}
		if _, was := s.touching[p]; !was {
			a.fireCollision(b)
			b.fireCollision(a)
			s.fireRules(a, b)
		}
	}
	s.touching, s.touchNext = now, s.touching
}

func (s *Scene) fireRules(a, b *Object) {
	for _, r := range s.rules {
		if a.tag == r.tagA && b.tag == r.tagB {
			r.fn(a, b)
		} else if a.tag == r.tagB && b.tag == r.tagA {
			r.fn(b, a)
		}
	}
}

// resolveSolids pushes non-solid objects out of solid ones along the
// axis of least penetration, and maintains Grounded for gravity objects.
func (s *Scene) resolveSolids() {
	solids := s.solidBuf[:0]
	for _, o := range s.objects {
		if !o.dead && o.solid {
			solids = append(solids, o)
		}
	}

	for _, o := range s.objects {
		if o.dead || o.solid || o.visual {
			continue
		}
		o.grounded = false
		for _, w := range solids {
			push, hit := physics.Resolve(o.box(), w.box())
			if !hit {
				continue
			}
			o.X += push.X
			o.Y += push.Y
			if push.Y < 0 {
				o.grounded = true
				if o.gravity && o.Vy > 0 {
					o.Vy = 0
				}
			}
			if push.Y > 0 && o.gravity && o.Vy < 0 {
				o.Vy = 0
			}
		}
	}
	clear(solids) // keep the buffer, not the objects
	s.solidBuf = solids[:0]
}

// flush removes destroyed objects at the end of the frame, so Destroy
// is always safe to call from inside any callback.
func (s *Scene) flush() {
	alive := s.objects[:0]
	for _, o := range s.objects {
		if !o.dead {
			alive = append(alive, o)
		}
	}
	// Let the garbage collector reclaim dropped objects.
	for i := len(alive); i < len(s.objects); i++ {
		s.objects[i] = nil
	}
	s.objects = alive
	s.index.invalidate() // indices shifted
}

// drawOrder lists the live objects bottom to top: by layer, and in
// scene order within a layer. The slice is reused between calls.
func (s *Scene) drawOrder() []*Object {
	s.order = s.order[:0]
	for _, o := range s.objects {
		if !o.dead {
			s.order = append(s.order, o)
		}
	}
	byLayer := func(a, b *Object) int { return a.layer - b.layer }
	if !slices.IsSortedFunc(s.order, byLayer) {
		slices.SortStableFunc(s.order, byLayer)
	}
	return s.order
}

// draw renders every live object bottom to top: world objects shifted
// by the camera, Fixed ones straight to the screen. Objects entirely
// outside the screen are skipped.
func (s *Scene) draw(screen *ebiten.Image) {
	vx, vy := s.view()
	w, h := s.game.Width(), s.game.Height()
	for _, o := range s.drawOrder() {
		ox, oy := vx, vy
		if o.fixed {
			ox, oy = 0, 0
		}
		if !o.onScreen(ox, oy, w, h) {
			continue
		}
		o.draw(screen, ox, oy)
	}
}

// onScreen reports whether the object's drawing may touch the w x h
// screen whose top-left is the view origin (ox, oy). The bound is
// conservative: a rotated object counts with the circle around its
// box, and text and buttons, which can draw past their box, are always
// drawn.
func (o *Object) onScreen(ox, oy, w, h float64) bool {
	if o.isText || o.isButton {
		return true
	}
	hw, hh := o.w/2, o.h/2
	if o.rotation != 0 {
		r := math.Hypot(hw, hh)
		hw, hh = r, r
	}
	const margin = 1 // rounding and filtering at the edges
	x, y := o.X-ox, o.Y-oy
	return x+hw+margin >= 0 && x-hw-margin <= w && y+hh+margin >= 0 && y-hh-margin <= h
}
