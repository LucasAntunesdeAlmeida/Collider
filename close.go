package collider

import "github.com/hajimehoshi/ebiten/v2"

// The window's own close request (its X button, Alt+F4, the dock or
// taskbar's Close, a launcher such as Steam asking the game to exit) goes
// through OnClose. Swappable so tests can stand in for a window.
var (
	windowBeingClosed       = ebiten.IsWindowBeingClosed
	setWindowClosingHandled = ebiten.SetWindowClosingHandled
)

// OnClose runs fn once when the player closes the window (the X button,
// Alt+F4, the taskbar, a launcher's exit), before the game ends: the
// moment to save progress or bank a run in progress. fn runs on the
// game loop, between two frames, like any other callback; then the game
// quits the way Quit does (the window closes and Run returns, saving
// the COLLIDER_RECORD GIF on the way). There is no veto: a close request
// always ends the game, since launchers and the OS expect it to.
//
// Only a close request runs fn: Quit is the game's own decision and
// does not. Headless runs (Step, COLLIDER_AGENT=mcp) have no window, so
// fn never runs there. In a browser a page is closed by its tab, not by
// the game: fn never runs either (save as you go; saves there are
// written at once). OnClose(nil) restores the default: the window
// closes at once. Callable any time, before or after Run.
func (g *Game) OnClose(fn func()) {
	g.onClose = fn
	setWindowClosingHandled(fn != nil)
}

// closeRequested runs the close handler once the window is asked to
// close, and reports that the game must end now.
func (g *Game) closeRequested() bool {
	if g.onClose == nil || g.headless || !windowBeingClosed() {
		return false
	}
	fn := g.onClose
	g.onClose = nil // once, even if fn panics or the loop runs again
	fn()
	g.quit = true
	return true
}
