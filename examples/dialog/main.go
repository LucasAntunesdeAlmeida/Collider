// Dialog: a visual-novel style conversation with drawn portraits, a
// framed dialog box, a typewriter effect with letter blips, and a menu.
// Click to advance; clicking mid-line reveals the rest instantly.
// It speaks the player's language: English or Portuguese.
//
// Run from this folder: cd examples/dialog && go run .
// Try a language: COLLIDER_LANG=pt-BR go run .
package main

import (
	"strings"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

const font = "fonts/pixel.ttf"

type line struct {
	who  string // portrait: sprites/<who>.png
	text string
}

// tongue is every string the player reads, in one language.
type tongue struct {
	title, sub, hint, listen, click, over, again, menu string
	names                                              map[string]string
	script                                             []line
}

var tongues = map[string]tongue{
	"en": {
		title: "THE CAVE HUMS", sub: "A CONVERSATION", hint: "CLICK TO ADVANCE",
		listen: "LISTEN", click: "CLICK", over: "THE CONVERSATION ENDS",
		again: "HEAR IT AGAIN", menu: "MENU",
		names: map[string]string{"elder": "Elder", "hero": "Hero"},
		script: []line{
			{"elder", "The cave hums tonight, traveler."},
			{"hero", "I heard it from the road. What sleeps down there?"},
			{"elder", "Not sleeps. Waits."},
			{"hero", "...then I will bring my sword."},
			{"elder", "Bring your courage. Swords rust."},
		},
	},
	"pt": {
		title: "A CAVERNA ZUMBE", sub: "UMA CONVERSA", hint: "CLIQUE PARA AVANÇAR",
		listen: "OUVIR", click: "CLIQUE", over: "FIM DA CONVERSA",
		again: "OUVIR DE NOVO", menu: "MENU",
		names: map[string]string{"elder": "Ancião", "hero": "Herói"},
		script: []line{
			{"elder", "A caverna zumbe esta noite, viajante."},
			{"hero", "Ouvi da estrada. O que dorme lá embaixo?"},
			{"elder", "Não dorme. Espera."},
			{"hero", "...então levarei minha espada."},
			{"elder", "Leve sua coragem. Espadas enferrujam."},
		},
	},
}

func main() {
	g := engine.New("Dialog", 800, 600)

	// The player's language ("pt-BR"): its first part picks the
	// strings, and anything unknown falls back to English.
	lang, _, _ := strings.Cut(g.Language(), "-")
	t, ok := tongues[lang]
	if !ok {
		t = tongues["en"]
	}
	script := t.script

	// --- Menu ---
	menu := g.Scene("menu")
	menu.Music("audios/theme.wav")
	menu.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())
	menu.Add(engine.Text(t.title).At(400, 150).Font(font).TextSize(32).TextColor(engine.Yellow))
	menu.Add(engine.Text(t.sub).At(400, 220).Font(font).TextSize(14))
	menu.Add(engine.Text(t.hint).At(400, 500).Font(font).TextSize(12))
	menu.Add(engine.Button(t.listen).At(400, 360).Font(font).TextSize(22).Color(engine.Green)).
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
	play.Add(engine.Text(t.click).At(740, 560).Font(font).TextSize(9).Visual())

	idx := 0
	shown := 0.0
	blipAt := 0
	show := func() {
		l := script[idx]
		name.SetText(strings.ToUpper(t.names[l.who]))
		portrait.SetSprite("sprites/" + l.who + ".png")
		shown = 0
		blipAt = 0
	}
	show()

	speech.OnUpdate(func(dt float64) {
		if idx >= len(script) {
			return // conversation over, waiting for the scene switch
		}
		// Count letters, not bytes: "ã" and "é" are two bytes each.
		letters := []rune(script[idx].text)
		shown += 30 * dt
		n := min(len(letters), int(shown))
		speech.SetText(string(letters[:n]))
		if n > blipAt {
			blipAt = n + 3
			if n < len(letters) {
				g.Sound("audios/blip.wav")
			}
		}
	})

	play.OnClick(func(_, _ float64) {
		letters := len([]rune(script[idx].text))
		if int(shown) < letters {
			shown = float64(letters)
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
	end.Add(engine.Text(t.over).At(400, 220).Font(font).TextSize(22).TextColor(engine.Yellow))
	end.Add(engine.Button(t.again).At(290, 400).Font(font).TextSize(16).Color(engine.Green)).
		OnClick(func() {
			idx = 0
			show()
			g.Restart("play")
		})
	end.Add(engine.Button(t.menu).At(540, 400).Font(font).TextSize(16)).
		OnClick(func() { g.Go("menu") })

	g.Run("menu")
}
