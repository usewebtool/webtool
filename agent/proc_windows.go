//go:build windows

package agent

import "syscall"

// sysProcAttr returns process attributes for detaching the daemon on Windows.
func sysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
