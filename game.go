package collider

import (
	"image"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"

	"github.com/LucasAntunesdeAlmeida/collider/internal/assets"
	"github.com/LucasAntunesdeAlmeida/collider/internal/record"
)

// Game owns the window, the scenes, the asset cache and the main loop.
type Game struct {
	title  string
	width  int
	height int

	scenes  map[string]*Scene
	current *Scene
	next    *Scene
	quit    bool

	assets      *assets.Cache
	musicPath   string
	musicPlayer *audio.Player

	rec    *record.Recorder
	recBuf []byte
}

// New creates a game with a window title and size in pixels.
func New(title string, width, height int) *Game {
	return &Game{
		title:  title,
		width:  width,
		height: height,
		scenes: map[string]*Scene{},
		assets: assets.NewCache(),
	}
}

// Scene returns the scene with this name, creating it on first use.
func (g *Game) Scene(name string) *Scene {
	if s, ok := g.scenes[name]; ok {
		return s
	}
	s := newScene(g, name)
	g.scenes[name] = s
	return s
}

// Go switches to another scene at the end of the current frame.
// The scene keeps its state; use Restart to reset it.
func (g *Game) Go(name string) {
	g.next = g.mustScene(name)
}

// Restart switches to a scene after rolling it back to its initial
// state: setup objects restored, runtime spawns and timers dropped.
func (g *Game) Restart(name string) {
	s := g.mustScene(name)
	s.restart()
	g.next = s
}

func (g *Game) mustScene(name string) *Scene {
	s, ok := g.scenes[name]
	if !ok {
		panic("collider: unknown scene " + name)
	}
	return s
}

// Sound plays a short effect, fire and forget.
func (g *Game) Sound(path string) {
	g.assets.PlaySound(path)
}

// Quit closes the window and returns from Run.
func (g *Game) Quit() {
	g.quit = true
}

// Fullscreen switches fullscreen on or off. Callable any time,
// including from an input handler for an F11 toggle.
func (g *Game) Fullscreen(on bool) {
	ebiten.SetFullscreen(on)
}

// Resizable lets the player resize the window; the game keeps its
// logical resolution and scales.
func (g *Game) Resizable(on bool) {
	mode := ebiten.WindowResizingModeDisabled
	if on {
		mode = ebiten.WindowResizingModeEnabled
	}
	ebiten.SetWindowResizingMode(mode)
}

// Icon sets the window icon from an image asset.
func (g *Game) Icon(path string) {
	img := g.assets.Image(path)
	ebiten.SetWindowIcon([]image.Image{img})
}

// Width returns the window width in pixels.
func (g *Game) Width() float64 { return float64(g.width) }

// Height returns the window height in pixels.
func (g *Game) Height() float64 { return float64(g.height) }

// Run starts the game on the given scene and blocks until the window
// closes or Quit is called. If the COLLIDER_RECORD environment variable
// is set to a file path, the session is saved there as an animated GIF.
func (g *Game) Run(name string) {
	g.Go(name)
	if path := os.Getenv("COLLIDER_RECORD"); path != "" {
		g.rec = record.New(path, g.width, g.height)
	}
	ebiten.SetWindowTitle(g.title)
	ebiten.SetWindowSize(g.width, g.height)
	if err := ebiten.RunGame(&runner{g}); err != nil {
		panic(err)
	}
	if g.rec != nil {
		g.rec.Save()
	}
}

// playMusic reacts to scene switches: keep playing if the music is the
// same file, otherwise stop and start the new one (or silence).
func (g *Game) playMusic(path string) {
	if path == g.musicPath {
		return
	}
	if g.musicPlayer != nil {
		g.musicPlayer.Close()
		g.musicPlayer = nil
	}
	g.musicPath = path
	if path != "" {
		g.musicPlayer = g.assets.PlayMusic(path)
	}
}

// runner adapts Game to ebiten.Game without exposing Update/Draw/Layout
// on the public API.
type runner struct {
	g *Game
}

func (r *runner) Update() error {
	g := r.g
	if g.quit {
		return ebiten.Termination
	}
	if g.next != nil {
		g.current = g.next
		g.next = nil
		g.current.activate()
		g.playMusic(g.current.musicPath)
	}
	if g.current == nil {
		return nil
	}
	g.current.update(1.0 / float64(ebiten.TPS()))
	return nil
}

func (r *runner) Draw(screen *ebiten.Image) {
	g := r.g
	if g.current != nil {
		g.current.draw(screen)
	}
	if g.rec != nil && g.rec.ShouldCapture() {
		if g.recBuf == nil {
			g.recBuf = make([]byte, 4*g.width*g.height)
		}
		screen.ReadPixels(g.recBuf)
		g.rec.Add(g.recBuf)
	}
}

func (r *runner) Layout(_, _ int) (int, int) {
	return r.g.width, r.g.height
}
