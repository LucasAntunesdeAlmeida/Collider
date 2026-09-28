package physics

import (
	"math"
	"slices"
)

// cellSize is the spatial hash cell edge. Objects larger than a cell
// simply occupy several cells, so the value only affects performance,
// never correctness.
const cellSize = 128.0

// Cell coordinates are packed into sortable integers: 21 bits per axis,
// biased so negative cells sort before positive ones, and 22 bits for
// the box index. Boxes beyond ±2^20 cells (about ±134 million pixels)
// are clamped to the edge cells, which only makes them extra
// candidates: the exact test still decides.
const (
	cellBits  = 21
	cellBias  = 1 << (cellBits - 1)
	cellMax   = 1<<cellBits - 1
	indexBits = 64 - 2*cellBits
	indexMask = 1<<indexBits - 1
	// MaxBoxes is how many boxes one Pairs call can index.
	MaxBoxes = 1 << indexBits
)

// span is the inclusive cell range a box covers.
type span struct{ x0, y0, x1, y1 int }

// Grid is the broad phase: a uniform grid that turns boxes into the
// index pairs that might overlap. It keeps its buffers between calls,
// so a scene that calls it every frame allocates nothing once warm.
// The zero value is ready to use; a Grid is not safe for concurrent
// use.
type Grid struct {
	spans   []span
	entries []uint64 // row-major cell key << indexBits | box index
	pairs   []uint64 // low index << 32 | high index
	out     [][2]int
}

func cellOf(v float64) int {
	c := math.Floor(v / cellSize)
	switch {
	case c < -cellBias || c != c:
		return 0
	case c > cellMax-cellBias:
		return cellMax
	}
	return int(c) + cellBias
}

// Pairs returns the index pairs of boxes that share at least one grid
// cell and so might overlap; callers run the exact Overlaps test on
// each. Every pair appears once, low index first, sorted by (low,
// high), so iteration is deterministic. The result is owned by the
// Grid and valid until the next call.
func (g *Grid) Pairs(boxes []Box) [][2]int {
	if len(boxes) > MaxBoxes {
		panic("physics: too many boxes for one grid")
	}
	g.spans = g.spans[:0]
	g.entries = g.entries[:0]
	for i, b := range boxes {
		s := span{
			x0: cellOf(b.X - b.W/2), x1: cellOf(b.X + b.W/2),
			y0: cellOf(b.Y - b.H/2), y1: cellOf(b.Y + b.H/2),
		}
		g.spans = append(g.spans, s)
		for cy := s.y0; cy <= s.y1; cy++ {
			for cx := s.x0; cx <= s.x1; cx++ {
				key := uint64(cy)<<cellBits | uint64(cx)
				g.entries = append(g.entries, key<<indexBits|uint64(i))
			}
		}
	}
	// One sort groups the boxes by cell, each cell's boxes by index.
	slices.Sort(g.entries)

	g.pairs = g.pairs[:0]
	for start := 0; start < len(g.entries); {
		key := g.entries[start] >> indexBits
		end := start + 1
		for end < len(g.entries) && g.entries[end]>>indexBits == key {
			end++
		}
		cx, cy := int(key&cellMax), int(key>>cellBits)
		for i := start; i < end; i++ {
			a := int(g.entries[i] & indexMask)
			sa := g.spans[a]
			for j := i + 1; j < end; j++ {
				b := int(g.entries[j] & indexMask)
				sb := g.spans[b]
				// Two boxes sharing several cells meet in each of
				// them; only the first shared cell (the top-left of
				// the overlap of their ranges) reports the pair, so
				// no dedupe set is needed.
				if max(sa.x0, sb.x0) != cx || max(sa.y0, sb.y0) != cy {
					continue
				}
				g.pairs = append(g.pairs, uint64(a)<<32|uint64(b))
			}
		}
		start = end
	}
	slices.Sort(g.pairs)

	g.out = g.out[:0]
	for _, p := range g.pairs {
		g.out = append(g.out, [2]int{int(p >> 32), int(p & math.MaxUint32)})
	}
	return g.out
}

// CandidatePairs is Pairs on a throwaway Grid, returning a slice the
// caller owns. Scenes keep a Grid instead, so the buffers are reused.
func CandidatePairs(boxes []Box) [][2]int {
	var g Grid
	return slices.Clone(g.Pairs(boxes))
}
