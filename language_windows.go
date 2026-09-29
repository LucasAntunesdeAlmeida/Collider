package collider

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// platformLanguage is the user's Windows display language (Settings >
// Time & language), the one the system's own menus speak, falling back
// to the user's locale (regional format) on systems without the MUI
// call.
func platformLanguage() string {
	if langs, err := windows.GetUserPreferredUILanguages(windows.MUI_LANGUAGE_NAME); err == nil {
		for _, l := range langs {
			if l != "" {
				return l
			}
		}
	}
	return userDefaultLocaleName()
}

var procGetUserDefaultLocaleName = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetUserDefaultLocaleName")

func userDefaultLocaleName() string {
	if procGetUserDefaultLocaleName.Find() != nil {
		return ""
	}
	const localeNameMaxLength = 85 // LOCALE_NAME_MAX_LENGTH
	var buf [localeNameMaxLength]uint16
	n, _, _ := procGetUserDefaultLocaleName.Call(uintptr(unsafe.Pointer(&buf[0])), localeNameMaxLength)
	if n == 0 {
		return ""
	}
	return windows.UTF16ToString(buf[:])
}
