//go:build !windows

package server

import "os/exec"

func setSysProcAttr(cmd *exec.Cmd) {
	// No special sysproc attributes needed on Unix
}
