package collider

import "testing"

func TestWindowsLocaleFallback(t *testing.T) {
	// The fallback for systems without the MUI call answers too.
	got := languageTag(userDefaultLocaleName())
	if got == "" {
		t.Fatal("GetUserDefaultLocaleName gave no locale")
	}
	t.Logf("user locale: %q", got)
}
