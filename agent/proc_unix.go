//go:build !windows

package agent

import "syscall"

// sysProcAttr returns process attributes for detaching the daemon on Unix.
func sysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
