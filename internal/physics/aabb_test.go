package physics

import "testing"

func TestOverlaps(t *testing.T) {
	cases := []struct {
		name string
		a, b Box
		want bool
	}{
		{"apart", Box{0, 0, 10, 10}, Box{100, 0, 10, 10}, false},
		{"overlapping", Box{0, 0, 10, 10}, Box{5, 0, 10, 10}, true},
		{"touching edges do not collide", Box{0, 0, 10, 10}, Box{10, 0, 10, 10}, false},
		{"contained", Box{0, 0, 100, 100}, Box{0, 0, 10, 10}, true},
		{"diagonal overlap", Box{0, 0, 10, 10}, Box{7, 7, 10, 10}, true},
		{"diagonal apart", Box{0, 0, 10, 10}, Box{11, 11, 10, 10}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Overlaps(c.a, c.b); got != c.want {
				t.Fatalf("Overlaps(%v, %v) = %v, want %v", c.a, c.b, got, c.want)
			}
			if got := Overlaps(c.b, c.a); got != c.want {
				t.Fatalf("Overlaps should be symmetric: (%v, %v) = %v, want %v", c.b, c.a, got, c.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	b := Box{X: 100, Y: 100, W: 50, H: 20}
	cases := []struct {
		name string
		x, y float64
		want bool
	}{
		{"center", 100, 100, true},
		{"left edge", 75, 100, true},
		{"right edge", 125, 100, true},
		{"outside x", 74, 100, false},
		{"outside y", 100, 111, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Contains(b, c.x, c.y); got != c.want {
				t.Fatalf("Contains(%v, %v, %v) = %v, want %v", b, c.x, c.y, got, c.want)
			}
		})
	}
}
