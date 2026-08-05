# Memory (point-and-click puzzle)

The counter-example: a complete game with **no movement and no physics at all**.
Clicks are point-vs-object collisions, so a click-driven game is still squarely
inside the engine's model, and it stays tiny.

The full game is in [main.go](main.go). Run it from this folder so the
assets resolve:

```bash
cd examples/memory
go run .
```

## What this example proves

- An entire genre carried by click events alone
- `After` for the wait-then-flip-back moment that nearly every puzzle game needs
- `Data` as the free per-object field for game state (the card's face)
- `SetSprite` swapping an object's image at runtime
