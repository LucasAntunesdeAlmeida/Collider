//go:build !windows && !js

package collider

import (
	"os"
	"os/exec"
	"runtime"
)

// platformLanguage reads the environment the Unix way (envLanguage).
// On macOS, apps opened from the Finder or the Dock get no LANG, so the
// first preferred language from the system settings comes first.
func platformLanguage() string {
	if runtime.GOOS == "darwin" {
		if out, err := exec.Command("/usr/bin/defaults", "read", "-g", "AppleLanguages").Output(); err == nil {
			if tag := appleLanguage(string(out)); tag != "" {
				return tag
			}
		}
	}
	return envLanguage(os.Getenv)
}
