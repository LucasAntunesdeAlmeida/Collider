package collider

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Line breaking for Wrap. The rules are a small, predictable subset of
// Unicode line breaking (UAX #14) that covers Latin, Cyrillic and CJK
// text:
//
//   - "\n" always breaks.
//   - A run of spaces is a break opportunity; the spaces are dropped
//     from the end of a wrapped line. Non-breaking spaces (U+00A0,
//     U+202F, U+2007) never break.
//   - Between two characters with no space, a break is allowed next to
//     any CJK character (ideographs, kana, CJK and full-width
//     punctuation), and after a hyphen inside a word ("sci-fi").
//   - Basic kinsoku: never before closing punctuation (，。、！？」』）
//     and the ASCII , . ! ? ) and the like, small kana, ー, 々, …), never
//     after opening punctuation (「『（【《 and ( [ {).
//   - A space before ! ? : ; » % (French spacing) or after « does not
//     break, so "Bonjour !" keeps its "!".
//   - A word wider than the whole line breaks between characters.

// wrapLines breaks s into lines no wider than width, as measured by
// advance (the width of a string in pixels). width <= 0 only splits at
// "\n", leaving every line as it is.
func wrapLines(s string, width float64, advance func(string) float64) []string {
	paras := strings.Split(s, "\n")
	for i, p := range paras {
		paras[i] = strings.TrimSuffix(p, "\r")
	}
	if width <= 0 {
		return paras
	}
	var lines []string
	for _, p := range paras {
		lines = wrapParagraph(lines, p, width, advance)
	}
	return lines
}

// piece is an unbreakable run of text and the breakable spaces after it.
type piece struct {
	word, space string
}

// pieces cuts a paragraph at its break opportunities.
func pieces(p string) []piece {
	var out []piece
	var word, space strings.Builder
	var last rune // the last character of word, 0 when word is empty
	flush := func() {
		out = append(out, piece{word.String(), space.String()})
		word.Reset()
		space.Reset()
		last = 0
	}
	for i := 0; i < len(p); {
		u := unitAt(p, i) // a character plus any combining marks
		r, _ := utf8.DecodeRuneInString(u)
		i += len(u)
		switch {
		case breakableSpace(r):
			space.WriteString(u)
			continue
		case space.Len() > 0:
			// After spaces: a break, unless French spacing glues the
			// spaces to the words around them.
			if noBreakAfterSpace(r) || (last != 0 && noBreakBeforeSpace(last)) {
				word.WriteString(space.String())
				space.Reset()
			} else {
				flush()
			}
		case last != 0 && canBreakBetween(last, r):
			flush()
		}
		word.WriteString(u)
		last = r
	}
	if word.Len() > 0 || space.Len() > 0 || len(out) == 0 {
		flush()
	}
	return out
}

// wrapParagraph fills lines greedily with the paragraph's pieces.
func wrapParagraph(lines []string, p string, width float64, advance func(string) float64) []string {
	var cur strings.Builder
	var curW float64
	started := false // cur holds something (even a paragraph's leading spaces)
	var pending string
	var pendingW float64
	emit := func() {
		lines = append(lines, cur.String())
		cur.Reset()
		curW, started, pending, pendingW = 0, false, "", 0
	}
	for _, pc := range pieces(p) {
		w := advance(pc.word)
		if started && curW+pendingW+w > width {
			if cur.Len() > 0 {
				emit()
			} else { // only a paragraph's leading spaces: drop them
				started, pending, pendingW = false, "", 0
			}
		}
		if started {
			cur.WriteString(pending)
			cur.WriteString(pc.word)
			curW += pendingW + w
		} else if w <= width {
			cur.WriteString(pc.word)
			curW = w
		} else {
			// A word wider than the line: break it between characters,
			// at least one per line.
			for _, u := range units(pc.word) {
				uw := advance(u)
				if cur.Len() > 0 && curW+uw > width {
					lines = append(lines, cur.String())
					cur.Reset()
					curW = 0
				}
				cur.WriteString(u)
				curW += uw
			}
		}
		started = true
		pending = pc.space
		pendingW = 0
		if pending != "" {
			pendingW = advance(pending)
		}
	}
	emit()
	return lines
}

// unitAt returns the character at byte i of s together with the
// combining marks, variation selectors and joiners that follow it, so
// a break never separates an accent from its letter.
func unitAt(s string, i int) string {
	_, n := utf8.DecodeRuneInString(s[i:])
	j := i + n
	for j < len(s) {
		r, m := utf8.DecodeRuneInString(s[j:])
		if !unicode.In(r, unicode.Mn, unicode.Me) && r != 0x200D && (r < 0xFE00 || r > 0xFE0F) {
			break
		}
		j += m
	}
	return s[i:j]
}

func units(s string) []string {
	var out []string
	for i := 0; i < len(s); {
		u := unitAt(s, i)
		out = append(out, u)
		i += len(u)
	}
	return out
}

func breakableSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// canBreakBetween reports whether a line may break between two adjacent
// characters with no space between them.
func canBreakBetween(a, b rune) bool {
	switch {
	case closing(b) || opening(a):
		return false
	case cjk(a) || cjk(b):
		return true
	case (a == '-' || a == '‐') && unicode.IsLetter(b):
		return true // after the hyphen of "sci-fi", "Lebens-mittel"
	}
	return false
}

// cjk is a character that allows a break on either side: ideographs,
// kana, CJK symbols and punctuation, full-width forms.
func cjk(r rune) bool {
	switch {
	case r >= 0x2E80 && r <= 0x2FFF, // radicals, ideographic description
		r >= 0x3000 && r <= 0x303F,   // CJK symbols and punctuation
		r >= 0x3040 && r <= 0x30FF,   // hiragana, katakana
		r >= 0x3100 && r <= 0x312F,   // bopomofo
		r >= 0x31A0 && r <= 0x31FF,   // bopomofo extended, strokes, katakana ext
		r >= 0x3200 && r <= 0x4DBF,   // enclosed CJK, compatibility, ext A
		r >= 0x4E00 && r <= 0x9FFF,   // unified ideographs
		r >= 0xF900 && r <= 0xFAFF,   // compatibility ideographs
		r >= 0xFE30 && r <= 0xFE4F,   // CJK compatibility forms
		r >= 0xFF00 && r <= 0xFFEF,   // half-width and full-width forms
		r >= 0x20000 && r <= 0x3FFFF: // ideographs, planes 2 and 3
		return true
	}
	return false
}

// closing characters never start a line.
func closing(r rune) bool {
	return strings.ContainsRune(
		`,.!?:;%)]}'"»›`+ // ASCII and Latin
			`、。，．：；！？）］｝」』】〕〗〙〛》〉”’〟｡｣､･`+ // CJK closing punctuation
			`…‥ー々〻ゝゞヽヾ・`+ // ellipses, prolonged sound, iteration marks
			`ぁぃぅぇぉっゃゅょゎゕゖァィゥェォッャュョヮヵヶ`+ // small kana
			`ㇰㇱㇲㇳㇴㇵㇶㇷㇸㇹㇺㇻㇼㇽㇾㇿ`, r)
}

// opening characters never end a line.
func opening(r rune) bool {
	return strings.ContainsRune(`([{«‹`+`（［｛「『【〔〖〘〚《〈“‘〝｢`, r)
}

// French spacing: a space before these, or after the opening
// guillemets, belongs to the words around it.
func noBreakAfterSpace(r rune) bool  { return strings.ContainsRune(`!?:;»›%`, r) }
func noBreakBeforeSpace(r rune) bool { return r == '«' || r == '‹' }
