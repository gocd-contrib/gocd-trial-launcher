package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"archive/zip"
	"io"
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

func Unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}

	defer r.Close()

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			os.MkdirAll(f.Name, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(f.Name), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(f.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}