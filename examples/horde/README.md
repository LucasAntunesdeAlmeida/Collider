# Horde (a world bigger than the window)

![demo](demo.gif)

Survive the night on a field far bigger than the screen. WASD or arrows
move the hero, who throws a dart the way they face every 0.3 seconds;
chasers pour in from just beyond the edges of the view, wherever the
hero goes, and fall after three darts. One touch and the horde wins.

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

## Layers: draw order that survives spawning

Chasers spawn all night long, after the hero and the HUD exist. By
scene order alone they would draw over both. `Object.Layer(n)` fixes
the order once: scenery 0, chasers 1, darts 2, hero 3, HUD 4. Lower
layers draw first, objects on one layer keep the order they were added
in, and clicks go to whatever is drawn on top.

## Drawing effects: one sprite, many looks

```go
hero.FlipX(dx < 0)                                      // face the way you walk
play.Add(engine.Sprite("sprites/dart.png").Rotate(facing)) // point along the flight
chaser.Tint(engine.Orange)                              // a variant without new art
chaser.Flash(engine.White, 0.08)                        // the hit flash
chaser.Alpha(fade)                                      // fade out when fallen
```

- `FlipX(on)` mirrors the drawing: the hero's side-view sprite faces
  left or right with its movement, and chasers turn toward the hero.
- `Rotate(rad)` turns the drawing around the center: each dart is
  drawn along the direction it flies. Rotation and flip are drawing
  only; the collider stays the upright box.
- `Tint(c)` multiplies the colors: one green chaser sprite comes in
  three variants (untinted, yellow, orange).
- `o.Flash(c, sec)` draws a solid silhouette for a moment: a scene
  rule `OnCollision("chaser", "dart", ...)` flashes the chaser white,
  knocks it back and plays a hit sound. Three hits and it re-tags to
  `"fallen"`, so it no longer kills the hero, then `Alpha` fades it
  out over 0.3 seconds before it is destroyed.
