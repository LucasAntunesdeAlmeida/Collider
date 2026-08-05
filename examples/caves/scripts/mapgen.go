// Package scripts holds the map generation for Caves: pure algorithms,
// no engine types. Generation is just Go code; the game turns the grid
// into tagged engine objects.
package scripts

import "math/rand/v2"

const (
	// Cols x Rows cells of Cell pixels fill the 800x600 window.
	Cols = 40
	Rows = 30
	Cell = 20.0
)

// Generate returns a cave grid (true = wall) using the classic cellular
// automata method: random fill, then smoothing votes where crowded cells
// become wall and lonely walls erode away.
func Generate() [][]bool {
	grid := make([][]bool, Rows)
	for y := range grid {
		grid[y] = make([]bool, Cols)
		for x := range grid[y] {
			border := x == 0 || y == 0 || x == Cols-1 || y == Rows-1
			grid[y][x] = border || rand.Float64() < 0.44
		}
	}
	for range 4 {
		grid = smooth(grid)
	}
	return grid
}

func smooth(grid [][]bool) [][]bool {
	out := make([][]bool, Rows)
	for y := range out {
		out[y] = make([]bool, Cols)
		for x := range out[y] {
			out[y][x] = wallNeighbors(grid, x, y) >= 5
		}
	}
	return out
}

// wallNeighbors counts the 8 surrounding cells; out-of-bounds counts as
// wall, which keeps the cave sealed at the borders.
func wallNeighbors(grid [][]bool, x, y int) int {
	n := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			xx, yy := x+dx, y+dy
			if xx < 0 || yy < 0 || xx >= Cols || yy >= Rows || grid[yy][xx] {
				n++
			}
		}
	}
	return n
}

// FindOpen returns an open cell as close to the map center as possible:
// the player's spawn point.
func FindOpen(grid [][]bool) (int, int) {
	cx, cy := Cols/2, Rows/2
	for r := range Rows {
		for y := cy - r; y <= cy+r; y++ {
			for x := cx - r; x <= cx+r; x++ {
				if x >= 0 && y >= 0 && x < Cols && y < Rows && !grid[y][x] {
					return x, y
				}
			}
		}
	}
	return 1, 1
}

// Reachable returns every open cell reachable from a start cell (BFS
// flood fill). Placing gems only on reachable cells keeps every
// generated cave winnable.
func Reachable(grid [][]bool, startX, startY int) [][2]int {
	seen := make([][]bool, Rows)
	for y := range seen {
		seen[y] = make([]bool, Cols)
	}
	seen[startY][startX] = true
	queue := [][2]int{{startX, startY}}
	var out [][2]int
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		out = append(out, c)
		for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			x, y := c[0]+d[0], c[1]+d[1]
			if x < 0 || y < 0 || x >= Cols || y >= Rows || grid[y][x] || seen[y][x] {
				continue
			}
			seen[y][x] = true
			queue = append(queue, [2]int{x, y})
		}
	}
	return out
}
