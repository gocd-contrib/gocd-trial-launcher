package utils

import (
	"fmt"
	"os"
)

func Out(f string, t ...interface{}) {
	if len(t) > 0 {
		fmt.Fprintf(os.Stdout, f+"\n", t...)
	} else {
		fmt.Fprint(os.Stdout, f+"\n")
	}
}

func Err(f string, t ...interface{}) {
	if len(t) > 0 {
		fmt.Fprintf(os.Stderr, f+"\n", t...)
	} else {
		fmt.Fprint(os.Stderr, f+"\n")
	}
}

// Exits with exitCode after printing message.
// Automatically selects STDOUT vs STDERR depending
// on value of exitCode
func Die(exitCode int, f string, t ...interface{}) {
	if exitCode != 0 {
		Err(f, t)
	} else {
		Out(f, t)
	}

	os.Exit(exitCode)
}
