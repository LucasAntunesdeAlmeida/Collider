# Publishing a Collider game

From "runs on my machine" to "here's the link". The
[ship example](examples/ship/) has all of this wired up.

## One self-contained .exe (Windows)

Embed your assets so the binary ships alone:

```go
import "embed"

//go:embed sprites audios
var content embed.FS

func main() {
	engine.UseAssets(content)
	// everything else stays exactly the same
}
```

Build without a console window popping up behind the game:

```powershell
go build -ldflags "-s -w -H=windowsgui" -o mygame.exe .
```

`-s -w` strips debug info (smaller file), `-H=windowsgui` hides the
console. The resulting .exe is the entire game: send it to anyone.

Window polish from code: `g.Resizable(true)`, `g.Fullscreen(true)`,
`g.Icon("sprites/icon.png")` (the title bar icon; giving the .exe file
itself an icon needs a resource tool like `goversioninfo`, optional).

## In the browser (WebAssembly) and itch.io

Every Collider game builds for the web:

```powershell
$env:GOOS="js"; $env:GOARCH="wasm"
go build -o web/game.wasm .
$env:GOOS=$null; $env:GOARCH=$null
copy "$(go env GOROOT)\lib\wasm\wasm_exec.js" web\
```

Add a `web/index.html`:

```html
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>My Game</title>
<style>body { margin: 0; background: #111; }</style>
<script src="wasm_exec.js"></script>
<script>
const go = new Go();
WebAssembly.instantiateStreaming(fetch("game.wasm"), go.importObject)
  .then(r => go.run(r.instance));
</script>
</head>
<body></body>
</html>
```

Test locally with any static server, then publish on itch.io: zip the
`web/` folder, create a new project, upload the zip, check "This file
will be played in the browser", set the viewport to your window size.
Your game is now a link.

## Linux

No cgo, no C compiler, no development packages: Linux builds
cross-compile from any machine, Windows included.

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o mygame .
```

The binary links only glibc and opens the rest when it starts, all
standard on a Linux desktop and in Valve's Steam Runtime: `libX11`
and its extensions (`libXrandr`, `libXcursor`, `libXi`,
`libXinerama`, `libXext`, `libXrender`), `libGL` (or `libEGL`), and
for sound a PulseAudio or PipeWire server (spoken to directly), else
`libasound`. A player whose machine has no working audio device
still gets the game: it plays silently and logs one line saying why.

## Third-party license notices

A shipped binary contains Ebitengine (Apache 2.0) and other permissive
dependencies. Ship a notices file next to your game:

```powershell
go install github.com/google/go-licenses@latest
go-licenses report . --template licenses.tpl > THIRD-PARTY-NOTICES.txt
```

(Or simply `go-licenses save . --save_path=third_party` to copy the
license files.) Everything Collider pulls in is MIT/BSD/Apache style;
attribution is the only obligation, closed-source and commercial use
are fine.
