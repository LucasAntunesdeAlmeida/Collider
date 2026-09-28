# Horde (a world bigger than the window)

![demo](demo.gif)

Survive the night on a field far bigger than the screen. WASD or arrows
move the hero; chasers pour in from just beyond the edges of the view,
wherever the hero goes. One touch and the horde wins.

```bash
cd examples/horde
go run .
```

## What this example proves: the camera

```go
play.Camera(hero.X, hero.Y)                                  // follow the hero
play.Add(engine.Text("0 S").At(400, 30).Fixed())             // HUD pinned to the screen
play.Add(engine.Rect(800, 600, engine.Black).At(400, 300).Fixed()) // backdrop too
```

- `Scene.Camera(x, y)` centers the view on a world point. Every object
  keeps plain world coordinates (the field spans -2000..2000 here) and
  the engine shifts the drawing. Call it every frame to follow the
  player; ease toward the target for a smoothed camera.
- `Object.Fixed()` pins an object to the screen: position in screen
  pixels, drawn and clicked where it appears, never colliding. The
  clock and the ground color are fixed; the scenery and the chasers
  scroll.
- Clicks follow the camera too: world objects are hit where they
  appear, and `Scene.OnClick` receives world coordinates.
