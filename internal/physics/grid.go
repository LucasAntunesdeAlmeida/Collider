package physics

import (
	"math"
	"slices"
)

// cellSize is the spatial hash cell edge. Objects larger than a cell
// simply occupy several cells, so the value only affects performance,
// never correctness.
const cellSize = 128.0

// CandidatePairs is the broad phase: it returns index pairs of boxes
// that share at least one grid cell and so might overlap. Callers run
// the exact Overlaps test on each candidate. Pairs are deduplicated,
// ordered (low index first) and sorted, so iteration is deterministic.
func CandidatePairs(boxes []Box) [][2]int {
	cells := map[[2]int][]int{}
	for i, b := range boxes {
		x0 := int(math.Floor((b.X - b.W/2) / cellSize))
		x1 := int(math.Floor((b.X + b.W/2) / cellSize))
		y0 := int(math.Floor((b.Y - b.H/2) / cellSize))
		y1 := int(math.Floor((b.Y + b.H/2) / cellSize))
		for cx := x0; cx <= x1; cx++ {
			for cy := y0; cy <= y1; cy++ {
				key := [2]int{cx, cy}
				cells[key] = append(cells[key], i)
			}
		}
	}

	seen := map[[2]int]struct{}{}
	var out [][2]int
	for _, members := range cells {
		for i := range members {
			for j := i + 1; j < len(members); j++ {
				a, b := members[i], members[j]
				if a > b {
					a, b = b, a
				}
				p := [2]int{a, b}
				if _, dup := seen[p]; dup {
					continue
				}
				seen[p] = struct{}{}
				out = append(out, p)
			}
		}
	}
	slices.SortFunc(out, func(a, b [2]int) int {
		if a[0] != b[0] {
			return a[0] - b[0]
		}
		return a[1] - b[1]
	})
	return out
}
