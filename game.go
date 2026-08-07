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

	input     inputSource
	headless  bool
	agentsOff bool
	agentIn   *agentInput
	pilot     func(Observation) Action
	controls  map[string]Key
	stateFn   func() any
	agentDocs string

	// Windowed MCP: commands cross from the stdio goroutine to the
	// game loop through agentReq, so all game state stays on one
	// thread; agentDone closes when the window does, so in-flight MCP
	// calls return instead of deadlocking.
	agentReq  chan agentCmd
	agentDone chan struct{}
	agentCur  *agentCmd
	agentHold Action

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
		input:  &realInput{},
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

// Sound plays a short effect, fire and forget. Silent in headless
// (agent) runs so bots and CI never touch the audio device.
func (g *Game) Sound(path string) {
	if g.headless {
		return
	}
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
//
// Agent play (unless the game called DisallowAgents):
//   - COLLIDER_AGENT=mcp runs headless as an MCP server on stdio; act
//     steps the simulation deterministically.
//   - COLLIDER_AGENT=mcp-window opens the window and runs in real time
//     while serving the same MCP tools: agents and the person at the
//     keyboard play together, and the session is watchable (and
//     recordable with COLLIDER_RECORD).
func (g *Game) Run(name string) {
	switch os.Getenv("COLLIDER_AGENT") {
	case "mcp":
		if !g.agentsOff {
			g.Headless(name)
			g.serveMCP()
			return
		}
	case "mcp-window":
		if !g.agentsOff {
			g.agentIn = &agentInput{keys: map[Key]bool{}}
			g.input = &mixedInput{agent: g.agentIn}
			g.agentReq = make(chan agentCmd)
			g.agentDone = make(chan struct{})
			go g.serveMCP()
		}
	}
	g.Go(name)
	if path := os.Getenv("COLLIDER_RECORD"); path != "" {
		g.rec = record.New(path, g.width, g.height)
	}
	ebiten.SetWindowTitle(g.title)
	ebiten.SetWindowSize(g.width, g.height)
	if err := ebiten.RunGame(&runner{g}); err != nil {
		panic(err)
	}
	if g.agentDone != nil {
		close(g.agentDone)
	}
	if g.rec != nil {
		g.rec.Save()
	}
}

// advance is one frame of game logic: scene switches, then the current
// scene's update. Shared by the window loop and headless Step.
func (g *Game) advance(dt float64) {
	if g.next != nil {
		g.current = g.next
		g.next = nil
		g.current.activate()
		if !g.headless {
			g.playMusic(g.current.musicPath)
		}
	}
	if g.current != nil {
		g.current.update(dt)
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
	if g.pilot != nil && g.current != nil {
		g.injectAction(g.pilot(g.Observe()))
	}
	if g.agentReq != nil {
		g.serviceAgent()
	}
	g.advance(1.0 / float64(ebiten.TPS()))
	if g.agentReq != nil && g.agentCur != nil {
		g.agentCur.frames--
		if g.agentCur.frames <= 0 {
			g.agentCur.reply <- g.Observe()
			g.agentCur = nil
		}
	}
	if g.agentIn != nil {
		g.agentIn.click = false
	}
	return nil
}

// serviceAgent runs the windowed-MCP command queue on the game thread:
// dequeue at most one pending command, apply it, and keep the agent's
// last held input pressed between commands (its virtual controller).
func (g *Game) serviceAgent() {
	if g.agentCur == nil {
		select {
		case cmd := <-g.agentReq:
			g.agentCur = &cmd
			if cmd.reset != "" {
				g.Restart(cmd.reset)
			}
			if !cmd.observe {
				g.agentHold = cmd.action
			}
		default:
		}
	}
	g.injectAction(g.agentHold)
	g.agentHold.Click = false // click fires on one frame only
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
