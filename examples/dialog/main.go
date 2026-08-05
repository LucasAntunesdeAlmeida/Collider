// Dialog: a visual-novel style conversation with drawn portraits, a
// framed dialog box, a typewriter effect with letter blips, and a menu.
// Click to advance; clicking mid-line reveals the rest instantly.
//
// Run from this folder: cd examples/dialog && go run .
package main

import (
	"strings"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

const font = "fonts/pixel.ttf"

type line struct {
	who  string
	text string
}

var script = []line{
	{"Elder", "The cave hums tonight, traveler."},
	{"Hero", "I heard it from the road. What sleeps down there?"},
	{"Elder", "Not sleeps. Waits."},
	{"Hero", "...then I will bring my sword."},
	{"Elder", "Bring your courage. Swords rust."},
}

func main() {
	g := engine.New("Dialog", 800, 600)

	// --- Menu ---
	menu := g.Scene("menu")
	menu.Music("audios/theme.wav")
	menu.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	menu.Add(engine.Text("THE CAVE HUMS").At(400, 150).Font(font).TextSize(32).TextColor(engine.Yellow))
	menu.Add(engine.Text("A CONVERSATION").At(400, 220).Font(font).TextSize(14))
	menu.Add(engine.Text("CLICK TO ADVANCE").At(400, 500).Font(font).TextSize(12))
	menu.Add(engine.Button("LISTEN").At(400, 360).Font(font).TextSize(22).Color(engine.Green)).
		OnClick(func() { g.Restart("play") })

	// --- Conversation ---
	play := g.Scene("play")
	play.Music("audios/theme.wav")
	play.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())

	// The dialog frame: dark panel with a thin light border.
	play.Add(engine.Rect(784, 184, engine.White).At(400, 488).Visual())
	play.Add(engine.Rect(776, 176, engine.Black).At(400, 488).Visual())

	portrait := play.Add(engine.Sprite("sprites/elder.png").At(105, 470).Visual())
	name := play.Add(engine.Text("").At(300, 425).Font(font).TextSize(22).TextColor(engine.Yellow))
	speech := play.Add(engine.Text("").At(440, 495).Font(font).TextSize(13))
	play.Add(engine.Text("CLICK").At(740, 560).Font(font).TextSize(9).Visual())

	idx := 0
	shown := 0.0
	blipAt := 0
	show := func() {
		l := script[idx]
		name.SetText(strings.ToUpper(l.who))
		portrait.SetSprite("sprites/" + strings.ToLower(l.who) + ".png")
		shown = 0
		blipAt = 0
	}
	show()

	speech.OnUpdate(func(dt float64) {
		if idx >= len(script) {
			return // conversation over, waiting for the scene switch
		}
		l := script[idx]
		shown += 30 * dt
		n := min(len(l.text), int(shown))
		speech.SetText(l.text[:n])
		if n > blipAt {
			blipAt = n + 3
			if n < len(l.text) {
				g.Sound("audios/blip.wav")
			}
		}
	})

	play.OnClick(func(_, _ float64) {
		l := script[idx]
		if int(shown) < len(l.text) {
			shown = float64(len(l.text))
			return
		}
		idx++
		if idx >= len(script) {
			g.Go("end")
			return
		}
		show()
	})

	// --- End ---
	end := g.Scene("end")
	end.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	end.Add(engine.Text("THE CONVERSATION ENDS").At(400, 220).Font(font).TextSize(22).TextColor(engine.Yellow))
	end.Add(engine.Button("HEAR IT AGAIN").At(290, 400).Font(font).TextSize(16).Color(engine.Green)).
		OnClick(func() {
			idx = 0
			show()
			g.Restart("play")
		})
	end.Add(engine.Button("MENU").At(540, 400).Font(font).TextSize(16)).
		OnClick(func() { g.Go("menu") })

	g.Run("menu")
}
