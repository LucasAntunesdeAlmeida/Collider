# Pong (arcade)

Two paddles, a ball, a score. The ball moves itself: objects have built-in
velocity (`Vx`, `Vy`) applied every frame, so constant motion costs zero lines.

The full game is in [main.go](main.go), about 40 lines. Run it from this
folder so the sound file resolves:

```bash
cd examples/pong
go run .
```

## What this example proves

- Built-in velocity: the ball needs no `OnUpdate` just to move in a straight line
- `Solid()` paddles interacting with a moving object
- `Text` objects and `SetText` for the score
- Fire-and-forget sound effects with `g.Sound`
