package collider

import (
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

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

	timers   []*timer
	onClick  []func(x, y float64)
	onUpdate []func(dt float64)
	rules    []collisionRule

	// touching tracks currently overlapping pairs so collision events
	// fire on enter, not on every frame of overlap.
	touching map[pair]struct{}
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
		game:     g,
		name:     name,
		touching: map[pair]struct{}{},
	}
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
// object-level click handling.
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
	s.initial = slices.Clone(s.objects)
	for _, o := range s.initial {
		o.saveState()
	}
}

// restart rolls the scene back to its initial snapshot: setup objects
// restored (including destroyed ones), runtime spawns dropped, initial
// timers re-armed, runtime timers dropped.
func (s *Scene) restart() {
	if !s.activated {
		return
	}
	s.objects = slices.Clone(s.initial)
	for _, o := range s.objects {
		o.restoreState()
	}
	s.touching = map[pair]struct{}{}
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
// object under the cursor.
func (s *Scene) clicks() {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}
	x, y := s.game.Mouse()
	for _, fn := range slices.Clone(s.onClick) {
		fn(x, y)
	}
	for _, o := range slices.Backward(s.objects) {
		if o.dead || len(o.onClick) == 0 {
			continue
		}
		if o.contains(x, y) {
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
	idx := make([]int, 0, len(s.objects))
	boxes := make([]physics.Box, 0, len(s.objects))
	for i, o := range s.objects {
		if o.dead || o.isText {
			continue
		}
		idx = append(idx, i)
		boxes = append(boxes, o.box())
	}

	now := map[pair]struct{}{}
	for _, c := range physics.CandidatePairs(boxes) {
		a := s.objects[idx[c[0]]]
		b := s.objects[idx[c[1]]]
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
	s.touching = now
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
	var solids []*Object
	for _, o := range s.objects {
		if !o.dead && o.solid {
			solids = append(solids, o)
		}
	}
	for _, o := range s.objects {
		if o.dead || o.solid || o.isText {
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
}

func (s *Scene) draw(screen *ebiten.Image) {
	for _, o := range s.objects {
		if !o.dead {
			o.draw(screen)
		}
	}
}
