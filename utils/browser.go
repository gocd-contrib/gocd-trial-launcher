package utils

import (
	"runtime"
)

func OpenUrlInBrowser(url string) {
	switch runtime.GOOS {
	case `darwin`:
		Out(`exec: open $url`)
	case `linux`:
		Out(`choose either xdg-,gnome-,kde-open or fall through to python or echo at the last resort`)
	case `windows`:
		Out(`exec: start $url`)
	default:
		Out(`Open your browser to %s`, url)
	}
}
