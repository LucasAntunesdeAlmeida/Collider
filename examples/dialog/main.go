// Dialog: a visual-novel style conversation. Click to advance; clicking
// mid-line reveals the rest of it instantly (the typewriter effect is a
// few lines of OnUpdate).
//
// This example proves the text system: a custom bold font loaded from a
// TTF file for speaker names, text sizes for hierarchy, and text colors.
//
// Run from this folder: cd examples/dialog && go run .
package main

import (
	"strings"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

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
	play := g.Scene("play")
	play.Music("audios/theme.wav")

	play.Add(engine.Rect(780, 180, engine.Black).At(400, 490))
	portrait := play.Add(engine.Sprite("sprites/elder.png").At(100, 470))
	name := play.Add(engine.Text("").At(260, 430).
		Font("fonts/gobold.ttf").TextSize(30).TextColor(engine.Yellow))
	speech := play.Add(engine.Text("").At(440, 500).TextSize(20))

	idx := 0
	shown := 0.0
	show := func() {
		l := script[idx]
		name.SetText(l.who)
		portrait.SetSprite("sprites/" + strings.ToLower(l.who) + ".png")
		shown = 0
	}
	show()

	speech.OnUpdate(func(dt float64) {
		if idx >= len(script) {
			return // conversation over, waiting for the scene switch
		}
		l := script[idx]
		shown += 40 * dt
		n := min(len(l.text), int(shown))
		speech.SetText(l.text[:n])
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

	end := g.Scene("end")
	end.Add(engine.Text("The conversation ends.").At(400, 270).
		Font("fonts/gobold.ttf").TextSize(32))
	end.Add(engine.Text("Click to hear it again.").At(400, 330).TextColor(engine.Yellow))
	end.OnClick(func(_, _ float64) {
		idx = 0
		show()
		g.Restart("play")
	})

	g.Run("play")
}
