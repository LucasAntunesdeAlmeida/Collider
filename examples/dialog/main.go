// Dialog: a visual-novel style conversation with drawn portraits, a
// framed dialog box, a typewriter effect with letter blips, and a menu.
// Click to advance; clicking mid-line reveals the rest instantly.
// It speaks the player's language: English, Portuguese or Chinese,
// with a CJK pixel font filling in the glyphs the Latin one lacks.
//
// Run from this folder: cd examples/dialog && go run .
// Try a language: COLLIDER_LANG=zh-CN go run . (or pt-BR)
package main

import (
	"strings"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

// Every text uses both fonts: the pixel font draws what it has (Latin,
// Cyrillic), the CJK font fills in the rest, and it is only loaded once
// a Chinese glyph shows up. Sizes are multiples of 12, where the 12px
// CJK font is sharp.
const (
	font = "fonts/pixel.ttf"
	cjk  = "fonts/cjk.ttf"
)

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
	"zh": {
		title: "洞穴在低鸣", sub: "一段对话", hint: "点击继续",
		listen: "倾听", click: "点击", over: "对话结束",
		again: "再听一次", menu: "菜单",
		names: map[string]string{"elder": "长老", "hero": "英雄"},
		script: []line{
			{"elder", "今晚洞穴在低鸣，旅人。"},
			{"hero", "我在路上就听到了那低鸣声。告诉我，下面究竟沉睡着什么？"},
			{"elder", "不是沉睡。是在等待。"},
			{"hero", "……那我就带上我的剑。"},
			{"elder", "带上你的勇气。剑会生锈。"},
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
	menu.Add(engine.Text(t.title).At(400, 150).Font(font, cjk).TextSize(48).TextColor(engine.Yellow))
	menu.Add(engine.Text(t.sub).At(400, 220).Font(font, cjk).TextSize(12))
	menu.Add(engine.Text(t.hint).At(400, 500).Font(font, cjk).TextSize(12))
	menu.Add(engine.Button(t.listen).At(400, 360).Font(font, cjk).TextSize(24).Color(engine.Green)).
		OnClick(func() { g.Restart("play") })

	// --- Conversation ---
	play := g.Scene("play")
	play.Music("audios/theme.wav")
	play.Add(engine.Sprite("sprites/background.png").At(400, 300).Visual())

	// The dialog frame: dark panel with a thin light border.
	play.Add(engine.Rect(784, 184, engine.White).At(400, 488).Visual())
	play.Add(engine.Rect(776, 176, engine.Black).At(400, 488).Visual())

	portrait := play.Add(engine.Sprite("sprites/elder.png").At(105, 470).Visual())
	// Speech wraps inside a fixed text area (Size): its lines start at
	// the area's top-left corner however many follow, in any language.
	// The name starts where the speech does: show re-anchors it by its
	// measured Width each time it changes.
	const left, width = 190, 560
	name := play.Add(engine.Text("").Font(font, cjk).TextSize(24).TextColor(engine.Yellow).At(left, 425))
	speech := play.Add(engine.Text("").Font(font, cjk).TextSize(24).
		Wrap(width).TextAlign(engine.AlignLeft).LineHeight(36).Size(width, 104).At(left+width/2, 502))
	// The hint ends at x 770 however long the word is in this language.
	hint := play.Add(engine.Text(t.click).Font(font, cjk).TextSize(12).Visual())
	hint.At(770-hint.Width()/2, 566)

	idx := 0
	shown := 0.0
	blipAt := 0
	show := func() {
		l := script[idx]
		name.SetText(strings.ToUpper(t.names[l.who]))
		name.X = left + name.Width()/2 // measured at once: left edge at x 190
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
	end.Add(engine.Text(t.over).At(400, 220).Font(font, cjk).TextSize(24).TextColor(engine.Yellow))
	end.Add(engine.Button(t.again).At(290, 400).Font(font, cjk).TextSize(12).Color(engine.Green)).
		OnClick(func() {
			idx = 0
			show()
			g.Restart("play")
		})
	end.Add(engine.Button(t.menu).At(540, 400).Font(font, cjk).TextSize(12)).
		OnClick(func() { g.Go("menu") })

	g.Run("menu")
}
