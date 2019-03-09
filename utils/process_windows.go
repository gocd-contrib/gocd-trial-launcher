package utils

import "os/exec"

func EnablePgid(cmd *exec.Cmd) {}

func KillPgid(cmd *exec.Cmd) error { return nil }
