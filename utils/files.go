package utils

import (
	"os"
	"os/exec"
	"path/filepath"
)

func BaseDir() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0])) // shouldn't get an error here
	return dir
}

func CommandExists(command string) bool {
	if _, err := exec.LookPath(command); err == nil {
		return true
	} else {
		return false
	}
}
