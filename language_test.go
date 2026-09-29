package collider

import (
	"sync"
	"testing"
)

func TestLanguageTag(t *testing.T) {
	cases := map[string]string{
		"pt_BR.UTF-8":       "pt-BR",
		"pt_BR.UTF-8@euro":  "pt-BR",
		"de_DE@euro":        "de-DE",
		"zh_CN.GB18030":     "zh-CN",
		"zh-hans-cn":        "zh-Hans-CN",
		"zh-Hant-TW":        "zh-Hant-TW",
		"EN-us":             "en-US",
		"en_US":             "en-US",
		"fr":                "fr",
		"es-419":            "es-419",
		"ru_RU.KOI8-R":      "ru-RU",
		" de-DE ":           "de-DE",
		"sr_RS.UTF-8@latin": "sr-RS",
		"C":                 "",
		"C.UTF-8":           "",
		"POSIX":             "",
		"":                  "",
		"und":               "",
		"123":               "",
		"x":                 "",
	}
	for in, want := range cases {
		if got := languageTag(in); got != want {
			t.Errorf("languageTag(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEnvLanguage(t *testing.T) {
	cases := []struct {
		env  map[string]string
		want string
	}{
		{map[string]string{}, ""},
		{map[string]string{"LANG": "pt_BR.UTF-8"}, "pt-BR"},
		{map[string]string{"LANG": "C"}, ""},
		{map[string]string{"LANG": "C.UTF-8", "LANGUAGE": "de"}, ""}, // gettext ignores LANGUAGE under C
		{map[string]string{"LANG": "en_US.UTF-8", "LC_MESSAGES": "ru_RU.UTF-8"}, "ru-RU"},
		{map[string]string{"LANG": "en_US.UTF-8", "LC_MESSAGES": "ru_RU.UTF-8", "LC_ALL": "fr_FR.UTF-8"}, "fr-FR"},
		{map[string]string{"LANG": "en_US.UTF-8", "LANGUAGE": "zh_CN:en_US:en"}, "zh-CN"},
		{map[string]string{"LANG": "es_ES.UTF-8", "LANGUAGE": ":"}, "es-ES"},
		{map[string]string{"LC_ALL": "POSIX", "LANG": "de_DE.UTF-8"}, ""}, // LC_ALL wins, even as POSIX
	}
	for _, c := range cases {
		got := envLanguage(func(k string) string { return c.env[k] })
		if got != c.want {
			t.Errorf("envLanguage(%v) = %q, want %q", c.env, got, c.want)
		}
	}
}

func TestAppleLanguage(t *testing.T) {
	out := "(\n    \"pt-BR\",\n    en\n)\n"
	if got := appleLanguage(out); got != "pt-BR" {
		t.Fatalf("appleLanguage = %q, want pt-BR", got)
	}
	if got := appleLanguage("(\n    en,\n    \"zh-Hans-CN\"\n)\n"); got != "en" {
		t.Fatalf("appleLanguage = %q, want en", got)
	}
	if got := appleLanguage(""); got != "" {
		t.Fatalf("appleLanguage of nothing = %q, want empty", got)
	}
}

// fakeSystem stands in for the platform's answer and makes the engine
// treat this as a player's run, not a test run.
func fakeSystem(t *testing.T, lang string) {
	t.Helper()
	oldSys, oldTest := systemLanguage, underTest
	systemLanguage = func() string { return lang }
	systemLanguageOnce = sync.Once{}
	underTest = func() bool { return false }
	t.Setenv("COLLIDER_AGENT", "")
	t.Setenv("COLLIDER_LANG", "")
	t.Cleanup(func() {
		systemLanguage, underTest = oldSys, oldTest
		systemLanguageOnce = sync.Once{}
	})
}

func TestLanguageFromTheSystem(t *testing.T) {
	fakeSystem(t, "pt_BR.UTF-8")
	g := New("lang-test", 800, 600)
	if got := g.Language(); got != "pt-BR" {
		t.Fatalf("Language = %q, want pt-BR", got)
	}
}

func TestLanguageIsEmptyForAgentsAndTests(t *testing.T) {
	fakeSystem(t, "zh-CN")

	// go test itself.
	underTest = func() bool { return true }
	if got := New("lang-test", 800, 600).Language(); got != "" {
		t.Fatalf("under go test Language = %q, want empty", got)
	}
	underTest = func() bool { return false }

	// Headless and Autopilot, from the moment they are called.
	g := New("lang-test", 800, 600)
	g.Scene("play")
	if got := g.Language(); got != "zh-CN" {
		t.Fatalf("before Headless Language = %q, want zh-CN", got)
	}
	g.Headless("play")
	if got := g.Language(); got != "" {
		t.Fatalf("headless Language = %q, want empty", got)
	}
	p := New("lang-test", 800, 600)
	p.Autopilot(func(Observation) Action { return Action{} })
	if got := p.Language(); got != "" {
		t.Fatalf("autopilot Language = %q, want empty", got)
	}

	// COLLIDER_AGENT=mcp, but not once agents are disallowed, and not
	// in the windowed session a person plays along with.
	t.Setenv("COLLIDER_AGENT", "mcp")
	m := New("lang-test", 800, 600)
	if got := m.Language(); got != "" {
		t.Fatalf("mcp Language = %q, want empty", got)
	}
	m.DisallowAgents()
	if got := m.Language(); got != "zh-CN" {
		t.Fatalf("mcp with agents disallowed Language = %q, want zh-CN", got)
	}
	t.Setenv("COLLIDER_AGENT", "mcp-window")
	if got := New("lang-test", 800, 600).Language(); got != "zh-CN" {
		t.Fatalf("mcp-window Language = %q, want zh-CN", got)
	}
}

func TestColliderLangOverrides(t *testing.T) {
	fakeSystem(t, "en-US")
	t.Setenv("COLLIDER_LANG", "ru_RU.UTF-8")
	g := New("lang-test", 800, 600)
	if got := g.Language(); got != "ru-RU" {
		t.Fatalf("Language = %q, want ru-RU", got)
	}
	g.Scene("play")
	g.Headless("play") // agent runs too: that is how a bot tests a translation
	if got := g.Language(); got != "ru-RU" {
		t.Fatalf("headless Language = %q, want ru-RU", got)
	}
	g.DisallowAgents() // a player preference, like LANG
	if got := g.Language(); got != "ru-RU" {
		t.Fatalf("Language after DisallowAgents = %q, want ru-RU", got)
	}
	t.Setenv("COLLIDER_LANG", "C")
	if got := g.Language(); got != "" {
		t.Fatalf("COLLIDER_LANG=C Language = %q, want empty", got)
	}
}

func TestSystemLanguageIsATagOrEmpty(t *testing.T) {
	// Whatever this machine speaks, the real platform call answers
	// with a well-formed tag (or nothing) and never panics.
	got := languageTag(platformLanguage())
	if got != languageTag(got) {
		t.Fatalf("platform language %q does not normalize to itself", got)
	}
	t.Logf("this machine's language: %q", got)
}
