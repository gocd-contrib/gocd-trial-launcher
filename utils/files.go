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
	Debug(`Searching PATH for command %q`, command)
	if _, err := exec.LookPath(command); err == nil {
		Debug(`  Found.`)
		return true
	} else {
		Debug(`  No such command.`)
		return false
	}
}

func IsExist(path string) bool {
	Debug(`Checking if file %q exists`, path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		Debug(`  No.`)
		return false
	}

	Debug(`  Yes.`)
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
			Debug(`Checking if dir %q exists`, path)
			if !IsDir(path) {
				Debug(`  No.`)
				return false
			}
			Debug(`  Yes.`)
		}
	}
	return true
}

func MkdirP(paths ...string) error {
	for _, path := range paths {
		Debug(`Ensuring directory %q`, path)
		if err := os.MkdirAll(path, 0755); err != nil {
			Debug(`  Failed: %s`, err.Error())
			return err
		}
		Debug(`  Ok.`)
	}
	return nil
}
