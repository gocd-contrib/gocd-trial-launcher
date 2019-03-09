package utils

import (
	"os"
	"os/exec"
	"runtime"
)

func OpenUrlInBrowser(url string) {
	Out("\n\nGoCD server is available at: %s", url)
	if err := openUrl(url); err != nil {
		Out("\nUnable to launch your default browser: %v", err)
		Out("\nPlease open your browser to: %s", url)
	}
}

func openUrl(url string) error {
	switch runtime.GOOS {
	case `darwin`:
		return run(`open`, url)
	case `linux`:
		if CommandExists(`xdg-open`) {
			return run(`xdg-open`, url)
		} else if CommandExists(`gnome-open`) {
			return run(`gnome-open`, url)
		} else if CommandExists(`kde-open`) {
			return run(`kde-open`, url)
		} else if CommandExists(`python`) {
			return run(`python`, `-m`, `webbrowser`, url)
		} else {
			Out(`Open your browser to: %s`, url)
		}
	case `windows`:
		return run(`start`, url)
	default:
		Out(`Open your browser to: %s`, url)
	}

	return nil
}

func run(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Env = os.Environ()

	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Wait()
}
