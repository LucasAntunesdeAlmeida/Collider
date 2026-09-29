# Dialog (visual novel style, custom fonts)

![demo](demo.gif)

A short conversation with a typewriter effect. Click to advance; clicking
mid-line reveals the rest instantly. It speaks the player's language,
English or Portuguese.

```bash
cd examples/dialog
go run .
COLLIDER_LANG=pt-BR go run .   # try the Portuguese text on any system
```

## What this example proves

The text system with visual hierarchy:

```go
name := play.Add(engine.Text("").At(300, 425).
    Font("fonts/pixel.ttf").    // custom TTF, loaded by path and cached
    TextSize(22).               // larger than body text
    TextColor(engine.Yellow))   // colored speaker names
```

- **Custom fonts** load from a TTF/OTF file path; the default remains
  the embedded Go font, so simple games never need a font file.
- **`TextSize` and `TextColor`** give dialog its hierarchy: bold yellow
  names, plain white lines.
- **The typewriter effect is not an engine feature**: it is four lines
  in `OnUpdate` slicing the string by elapsed time, plus a click rule.
  Fewest lines, plain Go.
- **The player's language** comes from `g.Language()`, a tag such as
  `"pt-BR"` (Windows display language, macOS preferred language, `LANG`
  on Linux, `navigator.language` in a browser). Its first part picks
  the strings; anything else falls back to English:

  ```go
  lang, _, _ := strings.Cut(g.Language(), "-")
  t, ok := tongues[lang]
  if !ok {
      t = tongues["en"]
  }
  ```

  Agents and tests always get `""`, so they play in English on every
  machine; `COLLIDER_LANG` overrides it to try a translation. The
  typewriter counts letters (runes), not bytes, so "ã" never splits.
- `fonts/` joins `sprites/` and `audios/` in the example folder
  convention.
