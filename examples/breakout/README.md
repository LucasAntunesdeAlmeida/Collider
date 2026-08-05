# Breakout (brick breaker, with a real menu)

Grid spawning, a win condition, mouse-follow control, and the full scene flow
from menu to play to win or lose, all using the same objects and events model.

The full game is in [main.go](main.go). Run it from this folder so the
assets resolve:

```bash
cd examples/breakout
go run .
```

## What this example proves

- A real menu, built with the same objects and click events as gameplay
- Grid spawning with plain Go loops, no level format needed
- Mouse-follow control via `g.Mouse`
- Win and lose conditions with `Count` and screen-bounds checks
