package utils

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"runtime"
)

type JavaProps map[string]string

func (j *JavaProps) Args() []string {
	args := make([]string, 0)
	for k, v := range *j {
		args = append(args, `-D`+k+`=`+v)
	}
	return args
}

type Java struct {
	Home string
	java string
}

func (j *Java) Run(properties JavaProps, args ...string) *exec.Cmd {
	if len(properties) > 0 {
		args = append(properties.Args(), args...)
	}

	return exec.Command(j.java, args...)
}

func (j *Java) Verify() error {
	cmd := j.Run(nil, `version`)
	cmd.Stdout = ioutil.Discard
	cmd.Stderr = ioutil.Discard

	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Wait()
}

func (j *Java) Executable() string {
	return j.java
}

func NewJava(javaHome string) *Java {
	java := filepath.Join(javaHome, `bin`, `java`)

	if `windows` == runtime.GOOS {
		java += `.exe`
	}

	return &Java{Home: javaHome, java: java}
}
