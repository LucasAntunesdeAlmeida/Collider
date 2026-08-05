package physics

import (
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
