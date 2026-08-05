// Package physics holds collision primitives: AABB tests now, the
// spatial hash broad phase in M4.
package physics

// Box is an axis-aligned bounding box centered at X, Y.
type Box struct {
	X, Y, W, H float64
}

// Overlaps reports whether two boxes intersect.
func Overlaps(a, b Box) bool {
	dx := a.X - b.X
	if dx < 0 {
		dx = -dx
	}
	dy := a.Y - b.Y
	if dy < 0 {
		dy = -dy
	}
	return dx*2 < a.W+b.W && dy*2 < a.H+b.H
}

// Contains reports whether a point is inside the box, edges included.
func Contains(b Box, x, y float64) bool {
	return x >= b.X-b.W/2 && x <= b.X+b.W/2 &&
		y >= b.Y-b.H/2 && y <= b.Y+b.H/2
}

// Vec is a 2D displacement.
type Vec struct {
	X, Y float64
}

// Resolve returns the minimal translation that pushes box a out of box b,
// and whether the boxes overlapped at all. The push is along the axis of
// least penetration, which is what makes walls and floors feel right.
func Resolve(a, b Box) (Vec, bool) {
	dx := a.X - b.X
	if dx < 0 {
		dx = -dx
	}
	dy := a.Y - b.Y
	if dy < 0 {
		dy = -dy
	}
	px := (a.W+b.W)/2 - dx
	py := (a.H+b.H)/2 - dy
	if px <= 0 || py <= 0 {
		return Vec{}, false
	}
	if px < py {
		if a.X < b.X {
			return Vec{X: -px}, true
		}
		return Vec{X: px}, true
	}
	if a.Y < b.Y {
		return Vec{Y: -py}, true
	}
	return Vec{Y: py}, true
}
