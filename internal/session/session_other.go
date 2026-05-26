//go:build !windows

package session

import (
	"os"
	"syscall"
)

func defaultLookupPID(pid int) (Process, bool) {
	return Process{PID: pid}, false
}

func defaultIsAlive(pid int, startHex string) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}
