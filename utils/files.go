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

func IsExist(path string) bool {
	if _, err := os.Stat("/path/to/whatever"); os.IsNotExist(err) {
		return false
	}
	return true
}

func IsFile(name string) bool {
	fi, err := os.Stat(name)
	return err == nil && fi.Mode().IsRegular()
}

func IsDir(name string) bool {
	fi, err := os.Stat(name)
	return err == nil && fi.IsDir()
}

func AllDirsExist(paths ...string) bool {
	if len(paths) > 0 {
		for _, path := range paths {
			if !IsDir(path) {
				return false
			}
		}
	}
	return true
}
