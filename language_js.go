//go:build js

package collider

import "syscall/js"

// platformLanguage is the browser's language, navigator.language: the
// first of the user's preferred languages.
func platformLanguage() (lang string) {
	defer func() {
		if recover() != nil {
			lang = ""
		}
	}()
	nav := js.Global().Get("navigator")
	if !nav.Truthy() {
		return ""
	}
	v := nav.Get("language")
	if v.Type() != js.TypeString {
		return ""
	}
	return v.String()
}
