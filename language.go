package collider

import (
	"os"
	"strings"
	"sync"
)

// Language reports the player's language as a BCP 47 tag, such as
// "pt-BR", "zh-CN", "en-US" or "de", or "" when it is unknown: the
// moment to pick the game's translation, before or while building
// scenes (it needs only New). Match leniently: the primary subtag
// ("pt", "zh") picks the translation, the region or script only where
// it matters (zh-CN and zh-Hans are Simplified Chinese, zh-TW and
// zh-Hant Traditional), and anything unknown or "" falls back to the
// game's default.
//
// Where it comes from: on Windows, the user's display language; on
// macOS, the first preferred language (System Settings), then the
// environment; on Linux and other Unix systems, LC_ALL, LC_MESSAGES
// or LANG ("pt_BR.UTF-8" becomes "pt-BR"; "C" and "POSIX" mean
// unknown), with GNU's LANGUAGE list taking precedence as it does for
// gettext; in a browser, navigator.language.
//
// Agent and test runs (Headless, Autopilot, COLLIDER_AGENT=mcp, go
// test: the runs that are silent and keep saves in memory) get "", so
// the game plays in its default language and agents see the same text
// on every machine. Headless and Autopilot count from the moment they
// are called; a bot that builds the game first and then calls Headless
// should set COLLIDER_LANG.
//
// The COLLIDER_LANG environment variable overrides all of this, agent
// runs included, to try a translation without changing the system
// ("COLLIDER_LANG=zh-CN go run ."); COLLIDER_LANG=C forces "". Like
// LANG, it only picks among the game's own strings, so it is honored
// even after DisallowAgents.
func (g *Game) Language() string {
	if v := os.Getenv("COLLIDER_LANG"); v != "" {
		return languageTag(v)
	}
	if g.agentOrTestRun() {
		return ""
	}
	systemLanguageOnce.Do(func() { systemLanguageTag = languageTag(systemLanguage()) })
	return systemLanguageTag
}

// systemLanguage asks the platform (language_*.go); swappable so tests
// can stand in for a system. It is asked once per process.
var (
	systemLanguage     = platformLanguage
	systemLanguageOnce sync.Once
	systemLanguageTag  string
)

// languageTag turns a locale name as systems spell it into a BCP 47
// tag: "pt_BR.UTF-8@euro" is "pt-BR", "zh-hans-cn" is "zh-Hans-CN",
// "en_US" is "en-US". "C", "POSIX" and anything that does not start
// with a language code are unknown: "".
func languageTag(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexAny(s, ".@:"); i >= 0 {
		s = s[:i] // encoding, modifier, or the rest of a list
	}
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '_' || r == '-' })
	if len(parts) == 0 || !isLetters(parts[0]) || len(parts[0]) < 2 || len(parts[0]) > 8 {
		return "" // empty, "C", "POSIX", or not a locale at all
	}
	primary := strings.ToLower(parts[0])
	if primary == "posix" || primary == "und" {
		return ""
	}
	out := primary
	for _, p := range parts[1:] {
		switch {
		case len(p) == 4 && isLetters(p): // script: Hans, Latn
			p = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
		case len(p) == 2 && isLetters(p): // region: BR, CN
			p = strings.ToUpper(p)
		default: // numeric region (419), variants, extensions
			p = strings.ToLower(p)
		}
		out += "-" + p
	}
	return out
}

func isLetters(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return false
		}
	}
	return s != ""
}

// envLanguage is the Unix way: the first of LC_ALL, LC_MESSAGES and
// LANG that is set names the locale. When that locale is a real one
// (not unset, "C" or "POSIX"), GNU's LANGUAGE priority list wins, as
// it does for gettext.
func envLanguage(getenv func(string) string) string {
	var locale string
	for _, k := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := getenv(k); v != "" {
			locale = v
			break
		}
	}
	tag := languageTag(locale)
	if tag == "" {
		return ""
	}
	for _, l := range strings.Split(getenv("LANGUAGE"), ":") {
		if t := languageTag(l); t != "" {
			return t
		}
	}
	return tag
}

// appleLanguage reads the first entry of `defaults read -g
// AppleLanguages`, a property list array such as
//
//	(
//	    "pt-BR",
//	    en
//	)
func appleLanguage(out string) string {
	for _, line := range strings.Split(out, "\n") {
		line = strings.Trim(strings.TrimSpace(line), `(),"`)
		if t := languageTag(line); t != "" {
			return t
		}
	}
	return ""
}
