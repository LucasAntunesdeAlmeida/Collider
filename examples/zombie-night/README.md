# Zombie Night (top-down shooter)

![demo](demo.gif)

The stress test for **runtime spawning**. Objects are created and destroyed
constantly, so per-instance event wiring breaks down. Instead, collision rules
are declared once at the scene level, by tag pair, and apply to every current
and future object.

The full game is in [main.go](main.go). Run it from this folder so the
assets resolve:

```bash
cd examples/zombie-night
go run .
```

## What this example proves

- Scene-level collision rules by tag pair, the only wiring that survives
  constant spawning
- `Every` for spawn waves, without goroutines
- Helpers that kill boilerplate math: `AtEdge`, `MoveToward`, `VelocityToward`
- `LifeTime` auto-destroy, so bullets never leak
- This example is also why M4 brings the spatial-hash broad phase: dozens of
  moving objects make the naive pair scan the bottleneck
