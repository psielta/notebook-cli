//go:build windows

package session

import (
	"strconv"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

func defaultLookupPID(pid int) (Process, bool) {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return Process{PID: pid}, false
	}
	defer windows.CloseHandle(snap)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	err = windows.Process32First(snap, &entry)
	for err == nil {
		if int(entry.ProcessID) == pid {
			name := windows.UTF16ToString(entry.ExeFile[:])
			return Process{
				PID:       pid,
				ParentPID: int(entry.ParentProcessID),
				Name:      name,
				StartTime: processStartTime(uint32(pid)),
			}, true
		}
		err = windows.Process32Next(snap, &entry)
	}
	return Process{PID: pid}, false
}

func processStartTime(pid uint32) time.Time {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return time.Time{}
	}
	defer windows.CloseHandle(handle)

	var creation, exit, kernel, user windows.Filetime
	if err := windows.GetProcessTimes(handle, &creation, &exit, &kernel, &user); err != nil {
		return time.Time{}
	}
	return time.Unix(0, creation.Nanoseconds())
}

func defaultIsAlive(pid int, startHex string) bool {
	proc, ok := defaultLookupPID(pid)
	if !ok {
		return false
	}
	if startHex == "" {
		return true
	}
	if proc.StartTime.IsZero() {
		return false
	}
	return strconv.FormatInt(proc.StartTime.UnixNano(), 16) == startHex
}
