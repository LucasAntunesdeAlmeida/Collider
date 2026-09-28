package physics

import (
	"math"
	"math/rand/v2"
	"slices"
	"testing"
)

func hasPair(pairs [][2]int, a, b int) bool {
	return slices.Contains(pairs, [2]int{a, b})
}

func TestCandidatePairsFindsNeighbors(t *testing.T) {
	boxes := []Box{
		{X: 0, Y: 0, W: 10, H: 10},
		{X: 5, Y: 0, W: 10, H: 10},
		{X: 1000, Y: 1000, W: 10, H: 10},
	}
	pairs := CandidatePairs(boxes)
	if !hasPair(pairs, 0, 1) {
		t.Fatal("overlapping boxes must be candidates")
	}
	if hasPair(pairs, 0, 2) || hasPair(pairs, 1, 2) {
		t.Fatal("distant boxes must not be candidates")
	}
}

func TestCandidatePairsHandlesLargeBoxes(t *testing.T) {
	boxes := []Box{
		{X: 0, Y: 0, W: 400, H: 400}, // spans many cells
		{X: 180, Y: 0, W: 10, H: 10},
	}
	if !hasPair(CandidatePairs(boxes), 0, 1) {
		t.Fatal("a box larger than a cell must still meet its neighbors")
	}
}

func TestCandidatePairsNoDuplicates(t *testing.T) {
	// Two big boxes sharing many cells must appear exactly once.
	boxes := []Box{
		{X: 0, Y: 0, W: 400, H: 400},
		{X: 100, Y: 100, W: 400, H: 400},
	}
	n := 0
	for _, p := range CandidatePairs(boxes) {
		if p == [2]int{0, 1} {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("pair should appear exactly once, got %d", n)
	}
}

// sharesCell is the definition of a candidate: the two boxes' cell
// ranges intersect.
func sharesCell(a, b Box) bool {
	cell := func(v float64) int { return int(math.Floor(v / cellSize)) }
	return cell(a.X-a.W/2) <= cell(b.X+b.W/2) && cell(b.X-b.W/2) <= cell(a.X+a.W/2) &&
		cell(a.Y-a.H/2) <= cell(b.Y+b.H/2) && cell(b.Y-b.H/2) <= cell(a.Y+a.H/2)
}

func TestGridMatchesBruteForce(t *testing.T) {
	rng := rand.New(rand.NewPCG(7, 9))
	var g Grid
	for round := range 20 {
		n := 50 + rng.IntN(250)
		boxes := make([]Box, n)
		for i := range boxes {
			w, h := 4+rng.Float64()*60, 4+rng.Float64()*60
			if rng.IntN(20) == 0 {
				w, h = 300+rng.Float64()*500, 300+rng.Float64()*200 // spans many cells
			}
			boxes[i] = Box{X: (rng.Float64()*2 - 1) * 1500, Y: (rng.Float64()*2 - 1) * 1500, W: w, H: h}
		}
		var want [][2]int
		for i := range boxes {
			for j := i + 1; j < len(boxes); j++ {
				if sharesCell(boxes[i], boxes[j]) {
					want = append(want, [2]int{i, j})
				}
			}
		}
		// The reused grid and a fresh one agree with brute force: every
		// pair once, low index first, sorted.
		if got := g.Pairs(boxes); !slices.Equal(got, want) {
			t.Fatalf("round %d: reused grid gave %d pairs, want %d", round, len(got), len(want))
		}
		if got := CandidatePairs(boxes); !slices.Equal(got, want) {
			t.Fatalf("round %d: CandidatePairs gave %d pairs, want %d", round, len(got), len(want))
		}
	}
}

func TestGridFarAwayAndEmpty(t *testing.T) {
	var g Grid
	if len(g.Pairs(nil)) != 0 {
		t.Fatal("no boxes, no pairs")
	}
	// Beyond the packed cell range boxes clamp to the edge cells: extra
	// candidates at worst, never a missed pair.
	far := 1e12
	boxes := []Box{{X: far, Y: -far, W: 10, H: 10}, {X: far + 5, Y: -far, W: 10, H: 10}}
	if !hasPair(g.Pairs(boxes), 0, 1) {
		t.Fatal("overlapping boxes far away must still be candidates")
	}
}

func BenchmarkGridPairs(b *testing.B) {
	rng := rand.New(rand.NewPCG(1, 2))
	boxes := make([]Box, 700)
	for i := range boxes {
		boxes[i] = Box{X: (rng.Float64()*2 - 1) * 800, Y: (rng.Float64()*2 - 1) * 800, W: 32, H: 40}
	}
	var g Grid
	b.ReportAllocs()
	for range b.N {
		g.Pairs(boxes)
	}
}
