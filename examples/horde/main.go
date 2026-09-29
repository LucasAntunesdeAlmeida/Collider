// Horde: survive the night on a field far bigger than the window.
// WASD or arrows move the hero, who throws a dart every 0.3 seconds at
// the nearest chaser (or the way they face); chasers pour in from
// beyond the edges of the screen, wherever the hero goes, take three
// darts to fall, and wear the hero's health down while they touch.
//
// This example proves the camera: the scene follows the hero across the
// world with Camera, while the HUD and the ground color stay pinned to
// the screen with Fixed. Layers keep the drawing readable: scenery at
// the bottom, then chasers, darts, the hero, and the HUD on top,
// whatever order they spawn in. And it proves the drawing effects: the
// hero and chasers face their way with FlipX, darts fly Rotated along
// their path, chasers come in Tinted variants, Flash white when hit
// and fade out with Alpha when they fall. Area queries do the rest:
// Near aims the darts, Touching deals contact damage every moment a
// chaser is on the hero, not just when contact begins, and Hitbox keeps
// that contact fair: hero and chasers collide with their bodies, boxes
// smaller than their sprites, so a brushing arm or an empty sprite
// corner never costs health. Esc or P pauses: a "pause" scene opens as
// an Overlay on top of the frozen night, and so does switching to
// another window (Focused); its entries answer to the 1 and 2 keys,
// looked up with KeyNamed. On a touch screen, or with the mouse, press
// and drag anywhere to steer: MouseDown drives a virtual joystick. M
// mutes and unmutes everything with Volume. A gamepad plays it all: the
// left stick (PadAxis) or the d-pad (PadDown) walks, Start pauses, A
// and B pick the pause entries, and the menus show pad prompts while a
// pad is connected (PadConnected). The pause menu's QUIT (3, or Y)
// ends the game with Quit; in a browser, where a page cannot close its
// tab, CanQuit is false and the entry is not there.
//
// Run from this folder: cd examples/horde && go run .
package main

import (
	"fmt"
	"math"
	"math/rand/v2"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

const (
	font   = "fonts/pixel.ttf"
	field  = 2000.0 // the world spans -field..field on both axes
	maxHP  = 5
	barLen = 200.0
)

// Draw layers, bottom to top.
const (
	layerScenery = iota
	layerChasers
	layerDarts
	layerHero
	layerHUD
	layerStick
)

func main() {
	g := engine.New("Horde", 800, 600)

	// --- The night ---
	play := g.Scene("play")
	play.Music("audios/theme.wav")
	// The ground is pinned to the screen; the world scrolls over it.
	play.Add(engine.Rect(800, 600, engine.Black).At(400, 300).Fixed().Layer(layerScenery))
	scenery := []string{"sprites/tuft.png", "sprites/stone.png", "sprites/flower.png"}
	for range 500 {
		x, y := (rand.Float64()*2-1)*field, (rand.Float64()*2-1)*field
		play.Add(engine.Sprite(scenery[rand.IntN(len(scenery))]).At(x, y).Visual())
	}

	// Drawn 42x42, but the hero collides with its body only: the
	// sprite's empty corners and swinging arms never get hit.
	hero := play.Add(engine.Rect(42, 42, nil).At(0, 0).Tag("player").Layer(layerHero).
		Hitbox(24, 34).
		Animation("walk", "sprites/hero.png", 4, 10).
		Animation("idle", "sprites/idle.png", 1, 1))
	hero.Play("idle")
	clock := play.Add(engine.Text("0 S").At(400, 30).Font(font).TextSize(16).Fixed().Layer(layerHUD))
	tally := play.Add(engine.Text("KILLS 0").At(700, 30).Font(font).TextSize(16).Fixed().Layer(layerHUD))
	play.Add(engine.Rect(barLen+4, 16, engine.Black).At(130, 30).Fixed().Layer(layerHUD))
	bar := play.Add(engine.Rect(barLen, 12, engine.Red).At(130, 30).Fixed().Layer(layerHUD))

	// The virtual joystick: press anywhere and drag. The press point is
	// the center; the knob follows the pointer up to the ring's edge.
	ring := play.Add(engine.Sprite("sprites/ring.png").Fixed().Layer(layerStick).Alpha(0))
	knob := play.Add(engine.Sprite("sprites/knob.png").Fixed().Layer(layerStick).Alpha(0))
	held, sx, sy := false, 0.0, 0.0
	stick := func() (dx, dy float64) {
		if !g.MouseDown() {
			held = false
			ring.Alpha(0)
			knob.Alpha(0)
			return 0, 0
		}
		mx, my := g.Mouse()
		if !held { // just pressed: this point is the center
			held, sx, sy = true, mx, my
		}
		dx, dy = mx-sx, my-sy
		d := math.Hypot(dx, dy)
		if d > 48 { // the knob stops at the ring
			dx, dy = dx/d*48, dy/d*48
		}
		ring.At(sx, sy).Alpha(0.5)
		knob.At(sx+dx, sy+dy).Alpha(0.8)
		if d < 8 { // dead zone: a tap or a shaky thumb does not walk
			return 0, 0
		}
		return dx, dy
	}

	survived, kills, hp := 0.0, 0, maxHP
	facing := 0.0 // radians, the way the hero last moved
	safe := 0.0   // seconds of invulnerability left after a hit
	hero.OnUpdate(func(dt float64) {
		dx, dy := 0.0, 0.0
		if g.Key(engine.A) || g.Key(engine.Left) {
			dx--
		}
		if g.Key(engine.D) || g.Key(engine.Right) {
			dx++
		}
		if g.Key(engine.W) || g.Key(engine.Up) {
			dy--
		}
		if g.Key(engine.S) || g.Key(engine.Down) {
			dy++
		}
		// The d-pad walks like the keys.
		if g.PadDown(engine.PadLeft) {
			dx--
		}
		if g.PadDown(engine.PadRight) {
			dx++
		}
		if g.PadDown(engine.PadUp) {
			dy--
		}
		if g.PadDown(engine.PadDown) {
			dy++
		}
		speed := 1.0 // share of full speed
		if dx == 0 && dy == 0 {
			// The left stick is analog: a slight tilt walks slowly.
			// Axes are raw, so ignore a resting stick's small drift.
			sx, sy := g.PadAxis(engine.PadLeftX), g.PadAxis(engine.PadLeftY)
			if tilt := math.Hypot(sx, sy); tilt > 0.2 {
				dx, dy, speed = sx, sy, min(1, tilt)
			}
		}
		if sdx, sdy := stick(); dx == 0 && dy == 0 {
			dx, dy = sdx, sdy // no keys or pad: the drag joystick steers
		}
		if d := math.Hypot(dx, dy); d > 0 {
			hero.Move(dx/d*210*speed*dt, dy/d*210*speed*dt)
			hero.Play("walk")
			facing = math.Atan2(dy, dx)
			if dx != 0 {
				hero.FlipX(dx < 0) // the sprite faces right
			}
		} else {
			hero.Play("idle")
		}
		hero.X = max(-field, min(field, hero.X))
		hero.Y = max(-field, min(field, hero.Y))
		play.Camera(hero.X, hero.Y)

		survived += dt
		clock.SetText(fmt.Sprintf("%.0f S", survived))
	})

	// Darts fly at the nearest chaser in range, or the way the hero
	// faces, drawn turned to their path.
	play.Every(0.3, func() {
		aim := facing
		if near := play.Near("chaser", hero.X, hero.Y, 450); len(near) > 0 {
			aim = math.Atan2(near[0].Y-hero.Y, near[0].X-hero.X)
		}
		d := play.Add(engine.Sprite("sprites/dart.png").At(hero.X, hero.Y).
			Tag("dart").Layer(layerDarts).Rotate(aim))
		d.Vx, d.Vy = math.Cos(aim)*520, math.Sin(aim)*520
		d.LifeTime(1)
	})

	// Chasers appear on a ring just outside the view, around the hero,
	// in three tints of one sprite, with 3 hit points each.
	tints := []engine.Color{nil, engine.Yellow, engine.Orange}
	play.Every(0.6, func() {
		a := rand.Float64() * 2 * math.Pi
		c := play.Add(engine.Rect(36, 48, nil).
			At(hero.X+math.Cos(a)*560, hero.Y+math.Sin(a)*560).Tag("chaser").Layer(layerChasers).
			Hitbox(22, 40). // the body, not the outstretched arms
			Tint(tints[rand.IntN(len(tints))]).
			Animation("walk", "sprites/chaser.png", 2, 5))
		c.Play("walk")
		c.Data = 3
		speed := 70 + rand.Float64()*60
		fade := 1.0
		c.OnUpdate(func(dt float64) {
			if c.Data.(int) <= 0 { // fallen: fade out, then vanish
				fade -= dt / 0.3
				c.Alpha(fade)
				if fade <= 0 {
					c.Destroy()
				}
				return
			}
			c.FlipX(hero.X < c.X)
			c.MoveToward(hero.X, hero.Y, speed*dt)
		})
	})

	play.OnCollision("chaser", "dart", func(c, d *engine.Object) {
		d.Destroy()
		c.Flash(engine.White, 0.08)
		c.Move(d.Vx*0.03, d.Vy*0.03) // knocked back along the dart's path
		g.Sound("audios/hit.wav")
		c.Data = c.Data.(int) - 1
		if c.Data.(int) == 0 {
			c.Tag("fallen") // no longer deadly, no longer a target
			kills++
			tally.SetText(fmt.Sprintf("KILLS %d", kills))
		}
	})

	// g.Key and g.PadDown report a held button, and a press lasts
	// several frames, so toggling on "held" would flip the pause (or the
	// sound) every frame. edge remembers last frame's state and answers
	// true on the press only. Each checker is shared by both scenes,
	// since only one of them updates per frame.
	edge := func(held func() bool) func() bool {
		was := false
		return func() bool {
			down := held()
			hit := down && !was
			was = down
			return hit
		}
	}
	pausePressed := edge(func() bool {
		return g.Key(engine.Esc) || g.Key(engine.P) || g.PadDown(engine.PadStart)
	})
	mutePressed := edge(func() bool { return g.Key(engine.M) })
	padA := edge(func() bool { return g.PadDown(engine.PadA) })
	padB := edge(func() bool { return g.PadDown(engine.PadB) })
	padY := edge(func() bool { return g.PadDown(engine.PadY) })

	// prompt keeps a button's label in step with the controller in use:
	// the pad's button while a pad is connected, the key otherwise.
	prompt := func(b *engine.Object, keys, pad string) func() {
		shown := keys
		return func() {
			want := keys
			if g.PadConnected() {
				want = pad
			}
			if want != shown {
				shown = want
				b.SetText(want)
			}
		}
	}

	// M mutes and unmutes: sound effects and the music, at once.
	mutedLabel := play.Add(engine.Text("MUTED").At(700, 60).Font(font).TextSize(12).
		TextColor(engine.Yellow).Fixed().Layer(layerHUD).Alpha(0))
	muted := false
	muteKey := func() {
		if !mutePressed() {
			return
		}
		muted = !muted
		if muted {
			g.Volume(0)
			mutedLabel.Alpha(1)
		} else {
			g.Volume(1)
			mutedLabel.Alpha(0)
		}
	}

	play.OnUpdate(func(float64) {
		muteKey()
		// Esc, P or Start pauses; switching to another window pauses too.
		if pausePressed() || !g.Focused() {
			g.Overlay("pause")
		}
		// A does nothing here, but its edge is kept current, so an A
		// held when the night ends does not press AGAIN at once.
		padA()
	})

	// --- Game over ---
	over := g.Scene("over")

	over.Add(engine.Text("THE HORDE WON").At(400, 200).Font(font).TextSize(32).TextColor(engine.Red))
	result := over.Add(engine.Text("").At(400, 280).Font(font).TextSize(16))
	lose := func() {
		result.SetText(fmt.Sprintf("SURVIVED %.0f SECONDS, %d KILLS", survived, kills))
		g.Go("over")
	}
	again := func() {
		survived, kills, hp, facing, safe = 0, 0, maxHP, 0, 0
		g.Restart("play")
	}
	againBtn := over.Add(engine.Button("AGAIN").At(400, 420).Font(font).TextSize(18).Color(engine.Green))
	againBtn.OnClick(again)
	againPrompt := prompt(againBtn, "AGAIN", "A AGAIN")
	over.OnUpdate(func(float64) {
		againPrompt()
		if padA() {
			again()
		}
	})

	// Contact damage: every moment a chaser touches the hero, not just
	// the first, with a short blinking grace after each hit.
	hero.OnUpdate(func(dt float64) {
		safe -= dt
		if safe > 0 {
			hero.Alpha(0.35 + 0.65*float64(int(safe*20)%2))
			return
		}
		hero.Alpha(1)
		if len(hero.Touching("chaser")) == 0 {
			return
		}
		hp--
		safe = 0.5
		w := barLen * float64(hp) / maxHP
		bar.Size(w, 12).At(30+w/2, 30) // shrinks toward its left end
		g.Sound("audios/hurt.wav")
		if hp == 0 {
			lose()
		}
	})

	// --- Pause: an overlay. The night still draws underneath, frozen:
	// no chaser moves, no timer ticks, until it closes. ---
	pause := g.Scene("pause")
	pause.Add(engine.Rect(800, 600, engine.Black).At(400, 300).Alpha(0.6).Fixed())
	pause.Add(engine.Text("PAUSED").At(400, 170).Font(font).TextSize(32))
	resumeBtn := pause.Add(engine.Button("1 RESUME").At(400, 290).Font(font).TextSize(18).Color(engine.Green))
	resumeBtn.OnClick(g.CloseOverlay)
	giveUpBtn := pause.Add(engine.Button("2 GIVE UP").At(400, 380).Font(font).TextSize(18).Color(engine.Red))
	giveUpBtn.OnClick(lose) // Go closes the overlay too
	// QUIT closes the game after this frame, where a game can close: a
	// browser tab cannot be closed by its page, so there (CanQuit is
	// false) the entry is left out.
	quitPrompt := func() {}
	if g.CanQuit() {
		quitBtn := pause.Add(engine.Button("3 QUIT").At(400, 470).Font(font).TextSize(18))
		quitBtn.OnClick(g.Quit)
		quitPrompt = prompt(quitBtn, "3 QUIT", "Y QUIT")
	}
	// The number keys pick the entries too. The number row has no
	// constant; KeyNamed reaches any key by name, once, at setup. On a
	// gamepad, A resumes, B gives up and Y quits, and the labels say so.
	resumeKey, giveUpKey, quitKey := engine.KeyNamed("1"), engine.KeyNamed("2"), engine.KeyNamed("3")
	resumePrompt := prompt(resumeBtn, "1 RESUME", "A RESUME")
	giveUpPrompt := prompt(giveUpBtn, "2 GIVE UP", "B GIVE UP")
	pause.OnUpdate(func(float64) {
		muteKey()
		resumePrompt()
		giveUpPrompt()
		quitPrompt()
		// Read every edge every frame so none goes stale.
		resume, giveUp, quit := padA(), padB(), padY()
		if pausePressed() || g.Key(resumeKey) || resume {
			g.CloseOverlay()
		}
		if g.Key(giveUpKey) || giveUp {
			lose()
		}
		if g.Key(quitKey) || quit {
			g.Quit() // does nothing in a browser
		}
	})

	g.Run("play")
}
