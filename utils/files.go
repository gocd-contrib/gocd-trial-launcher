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

	_, err := exec.LookPath(command)
	return debugBoolUsing(err == nil, `  Found.`, `  No such command.`)
}

func IsExist(path string) bool {
	Debug(`Checking if file %q exists`, path)

	_, err := os.Stat(path)
	return debugBool(os.IsNotExist(err))
}

func IsFile(name string) bool {
	Debug(`Checking if %q is a file`, name)

	fi, err := os.Stat(name)
	return debugBool(err == nil && fi.Mode().IsRegular())
}

func IsDir(name string) bool {
	Debug(`Checking if %q is a directory`, name)

	fi, err := os.Stat(name)
	return debugBool(err == nil && fi.IsDir())
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

func debugBool(val bool) bool {
	return debugBoolUsing(val, `  Yes.`, `  No.`)
}

func debugBoolUsing(val bool, trueStr, falseStr string) bool {
	if val {
		Debug(trueStr)
	} else {
		Debug(falseStr)
	}
	return val
}
