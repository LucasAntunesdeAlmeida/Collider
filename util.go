package collider

import (
	"math/rand/v2"
	"slices"
)

// Shuffle returns a shuffled copy of a slice. Handy for card decks and
// spawn tables.
func Shuffle[T any](s []T) []T {
	out := slices.Clone(s)
	rand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}
