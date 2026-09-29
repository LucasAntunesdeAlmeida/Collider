# Dialog (visual novel style, custom fonts)

![demo](demo.gif)

A short conversation with a typewriter effect. Click to advance; clicking
mid-line reveals the rest instantly. It speaks the player's language:
English, Portuguese or Chinese.

```bash
cd examples/dialog
go run .
COLLIDER_LANG=zh-CN go run .   # try the Chinese text on any system (or pt-BR)
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
- **Fallback fonts** draw what the pixel font lacks. Every text names
  two fonts: the pixel font draws every glyph it has (Latin with
  accents, Cyrillic), and the CJK pixel font fills in the Chinese:

  ```go
  engine.Text(t.title).Font("fonts/pixel.ttf", "fonts/cjk.ttf")
  ```

  The CJK font is only read the first time a Chinese glyph shows up,
  so the English and Portuguese runs never load it, and text the pixel
  font draws alone lays out exactly as it did before. Mixed on one line,
  both fonts share the baseline. Pixel fonts are sharp at whole
  multiples of their design size, so the sizes here are multiples of 12
  (the CJK font's). `fonts/cjk.ttf` is a small subset of
  [Fusion Pixel Font](https://github.com/TakWolf/fusion-pixel-font)
  (SIL OFL, `fonts/cjk-OFL.txt`) holding only this example's Chinese
  glyphs; a real game ships the whole font (`zh_hans`, about 7 MB).
- **Wrapping** lays the speech out in the dialog box, in every
  language: English breaks at spaces, Chinese between any two
  characters (never before `，` or `。`):

  ```go
  speech := play.Add(engine.Text("").Font(font, cjk).TextSize(24).
      Wrap(560).TextAlign(engine.AlignLeft).LineHeight(36).
      Size(560, 104).At(190+280, 502))
  ```

  `Wrap(560)` makes the text a 560px column centered on its position,
  so `TextAlign(engine.AlignLeft)` starts every line at its left edge
  (x 190) whatever it says. `LineHeight(36)` spaces the lines (a pixel
  font's own line height has them touching). `Size` fixes the text
  area, so the first line stays put while the typewriter adds more. A
  word that no longer fits moves to the next line as it types, as in
  most games.
- **Measured text** lays out labels by their edges. A text's
  `Width()` and `Height()` are measured the moment it changes (fonts,
  fallback, size and wrap included), so the speaker's name starts at
  the speech's left edge, and the CLICK hint ends at the frame's right
  edge, whatever the language:

  ```go
  name.SetText(strings.ToUpper(t.names[l.who]))
  name.X = left + name.Width()/2
  ```

  The same call lets a test check that every translation fits its box.
- `fonts/` joins `sprites/` and `audios/` in the example folder
  convention.
