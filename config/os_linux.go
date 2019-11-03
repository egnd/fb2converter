// +build linux

package config

import (
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

// EnableColorOutput checks if colorized output is possible.
func EnableColorOutput(stream *os.File) bool {
	return terminal.IsTerminal(int(stream.Fd()))
}

// kindlegen provides OS specific part of default kindlegen location
func kindlegen() string {
	return "kindlegen"
}

// CleanFileName removes not allowed characters form file name.
func CleanFileName(in string) string {
	out := strings.TrimLeft(strings.Map(func(sym rune) rune {
		if strings.ContainsRune(string(os.PathSeparator)+string(os.PathListSeparator), sym) {
			return -1
		}
		return sym
	}, in), ".")
	if len(out) == 0 {
		out = "_bad_file_name_"
	}
	return out
}

// FindConverter  - used on Windows to support myhomelib
func FindConverter(_ string) string {
	return ""
}

// sqlite provides os specific part of default sqlite location
func sqlite() string {
	return "sqlite3"
}

// kpv returns os specific path where kindle previewer is installed by default.
func kpv() (string, error) {
	return "", ErrNoKPVForOS
}

// execute kpv - we need this as Windows requires special handling.
func kpvexec(exepath string, arg ...string) ([]string, error) {
	return nil, ErrNoKPVForOS
}
