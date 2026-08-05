# Catcher (save system)

Catch falling gems with the paddle for 30 seconds. Your best score is
saved to `save.json` and greets you on the next run.

```bash
cd examples/catcher
go run .
```

The paddle follows the mouse. Delete `save.json` to reset progress.

## What this example proves

**A save system needs nothing from the engine.** The whole persistence
layer is ~15 lines of plain Go:

```go
type saveData struct {
    Best int `json:"best"`
}

func load() saveData { /* os.ReadFile + json.Unmarshal, zero on first run */ }
func (s saveData) write() { /* json.MarshalIndent + os.WriteFile */ }
```

Design points worth copying:

- **Load once at startup, save at the moment that matters** (a new
  record), not every frame.
- **Extend `saveData` freely**: old save files keep loading, missing
  fields stay zero. That is versioning for free.
- **First run is not an error**: a missing file just returns zero values.

This settles the roadmap question of whether Collider needs `g.Save` /
`g.Load`: it does not. An engine wrapper would hide three standard
library calls without making any line count lower.
