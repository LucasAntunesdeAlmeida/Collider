package collider

import (
	"cmp"
	"math"
	"slices"

	"github.com/LucasAntunesdeAlmeida/collider/internal/physics"
)

// Area queries (Near, Touching) run against a per-scene spatial grid:
// one bucket per tag, holding each object keyed by the grid cell of its
// center, sorted so a row of cells is one contiguous run. A bucket is
// built lazily by the first query for its tag, and every bucket goes
// stale at the start of each frame and again after the motion step, so
// a frame builds at most twice per queried tag, however many queries
// it runs. Storage is reused from frame to frame.
//
// Lookups widen the searched area by queryMargin, and the final test
// always runs on current positions, so objects that moved a little
// since the build (a Move in an update callback) are still found
// exactly. Objects added since the build are scanned one by one. The
// one edge: an object teleported more than queryMargin within a frame,
// after the index was built, may only be seen by queries from the next
// update phase on.
const (
	queryCell   = 64.0
	queryMargin = 32.0
)

// queryIndex holds the tag buckets of one scene.
type queryIndex struct {
	buckets map[string]*tagBucket
	hits    []queryHit // scratch, reused between queries
}

// tagBucket indexes the live objects of one tag, as of its build.
type tagBucket struct {
	valid   bool
	n       int // len(scene.objects) at build time
	entries []cellEntry
	// halfW, halfH are the largest half extents in the bucket: an
	// object's box can reach that far from the cell holding its center.
	halfW, halfH float64
}

type cellEntry struct {
	key uint64
	i   int32 // index in scene.objects at build time
}

type queryHit struct {
	o    *Object
	i    int
	dist float64
}

// cellKey orders cells row by row, so cells cx0..cx1 of one row form
// one contiguous key range. The sign bit flip keeps negative
// coordinates in order.
func cellKey(cx, cy int32) uint64 {
	return uint64(uint32(cy)^1<<31)<<32 | uint64(uint32(cx)^1<<31)
}

// cellOf returns the grid cell of a world coordinate, clamped so any
// float (even infinite) maps to a valid cell.
func cellOf(v float64) int32 {
	c := math.Floor(v / queryCell)
	if !(c > math.MinInt32) { // also catches NaN
		return math.MinInt32
	}
	return int32(min(c, math.MaxInt32))
}

// invalidate marks every bucket stale; the next query rebuilds.
func (q *queryIndex) invalidate() {
	for _, b := range q.buckets {
		b.valid = false
	}
}

// bucket returns the up-to-date bucket for a tag, building it if needed.
func (s *Scene) bucket(tag string) *tagBucket {
	if s.index.buckets == nil {
		s.index.buckets = map[string]*tagBucket{}
	}
	b := s.index.buckets[tag]
	if b == nil {
		b = &tagBucket{}
		s.index.buckets[tag] = b
	}
	if b.valid {
		return b
	}
	b.valid = true
	b.n = len(s.objects)
	b.entries = b.entries[:0]
	b.halfW, b.halfH = 0, 0
	for i, o := range s.objects {
		if o.dead || o.tag != tag {
			continue
		}
		b.entries = append(b.entries, cellEntry{key: cellKey(cellOf(o.X), cellOf(o.Y)), i: int32(i)})
		b.halfW = max(b.halfW, o.w/2)
		b.halfH = max(b.halfH, o.h/2)
	}
	slices.SortFunc(b.entries, func(a, b cellEntry) int {
		return cmp.Or(cmp.Compare(a.key, b.key), cmp.Compare(a.i, b.i))
	})
	return b
}

// gather collects into the scratch hits every object with the tag whose
// center may lie within ex, ey of (x, y), and for which keep is true on
// current state. It visits the bucket's cells in range, then the
// objects added since the build.
func (s *Scene) gather(tag string, x, y, ex, ey float64, keep func(*Object) (float64, bool)) []queryHit {
	b := s.bucket(tag)
	hits := s.index.hits[:0]
	ex += b.halfW + queryMargin
	ey += b.halfH + queryMargin
	cx0, cx1 := cellOf(x-ex), cellOf(x+ex)
	cy1 := cellOf(y + ey)
	for cy := cellOf(y - ey); ; cy++ {
		lo, hi := cellKey(cx0, cy), cellKey(cx1, cy)
		j, _ := slices.BinarySearchFunc(b.entries, lo, func(e cellEntry, k uint64) int {
			return cmp.Compare(e.key, k)
		})
		for ; j < len(b.entries) && b.entries[j].key <= hi; j++ {
			i := int(b.entries[j].i)
			if i >= len(s.objects) {
				continue
			}
			if o := s.objects[i]; o.tag == tag {
				if d, ok := keep(o); ok {
					hits = append(hits, queryHit{o: o, i: i, dist: d})
				}
			}
		}
		if cy >= cy1 {
			break
		}
	}
	for i := b.n; i < len(s.objects); i++ {
		if o := s.objects[i]; o.tag == tag {
			if d, ok := keep(o); ok {
				hits = append(hits, queryHit{o: o, i: i, dist: d})
			}
		}
	}
	s.index.hits = hits
	return hits
}

// Near returns the live objects with this tag whose box overlaps the
// circle of radius r around (x, y), closest center first (ties keep
// scene order). Visual objects count, Fixed ones (screen space) never
// do. The auto-aim and pickup-magnet query: s.Near("enemy", p.X, p.Y,
// 300). The slice is the caller's to keep.
func (s *Scene) Near(tag string, x, y, r float64) []*Object {
	hits := s.gather(tag, x, y, r, r, func(o *Object) (float64, bool) {
		if o.dead || o.fixed {
			return 0, false
		}
		// Closest point of the box to the circle center.
		dx := max(math.Abs(o.X-x)-o.w/2, 0)
		dy := max(math.Abs(o.Y-y)-o.h/2, 0)
		if dx*dx+dy*dy > r*r {
			return 0, false
		}
		return math.Hypot(o.X-x, o.Y-y), true
	})
	slices.SortFunc(hits, func(a, b queryHit) int {
		return cmp.Or(cmp.Compare(a.dist, b.dist), cmp.Compare(a.i, b.i))
	})
	return hitObjects(hits)
}

// Touching returns the objects with this tag whose box overlaps this
// one right now, in scene order: continuous contact, for damage over
// time or standing in a zone, next to the enter-only OnCollision
// events. Like collisions, it ignores visual, text and Fixed objects
// (and returns nothing when called on one), dead objects, and the
// object itself. The slice is the caller's to keep.
func (o *Object) Touching(tag string) []*Object {
	s := o.scene
	if s == nil || o.dead || o.visual {
		return nil
	}
	box := o.box()
	hits := s.gather(tag, o.X, o.Y, o.w/2, o.h/2, func(other *Object) (float64, bool) {
		if other == o || other.dead || other.visual {
			return 0, false
		}
		return 0, physics.Overlaps(box, other.box())
	})
	slices.SortFunc(hits, func(a, b queryHit) int { return cmp.Compare(a.i, b.i) })
	return hitObjects(hits)
}

func hitObjects(hits []queryHit) []*Object {
	if len(hits) == 0 {
		return nil
	}
	out := make([]*Object, len(hits))
	for i, h := range hits {
		out[i] = h.o
	}
	clear(hits) // the scratch must not keep objects alive
	return out
}
