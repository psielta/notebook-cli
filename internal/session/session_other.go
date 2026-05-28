//go:build !windows

package session

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const psLookupTimeout = 500 * time.Millisecond

func defaultTerminalID() (string, bool) {
	files := []*os.File{os.Stdin, os.Stdout, os.Stderr}
	for _, file := range files {
		if id, ok := terminalIDForFile(file); ok {
			return id, true
		}
	}
	return "", false
}

func terminalIDForFile(file *os.File) (string, bool) {
	if file == nil {
		return "", false
	}

	info, err := file.Stat()
	if err != nil || info.Mode()&os.ModeCharDevice == 0 {
		return "", false
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "", false
	}

	value := strconv.FormatUint(uint64(stat.Dev), 16) + ":" + strconv.FormatUint(uint64(stat.Rdev), 16)
	return FormatDerivedSessionID("tty", value), true
}

func defaultLookupPID(pid int) (Process, bool) {
	if pid <= 0 {
		return Process{PID: pid}, false
	}
	if proc, ok := lookupProcfsPID(pid); ok {
		return proc, true
	}
	return lookupPSPID(pid)
}

func defaultIsAlive(pid int, startHex string) bool {
	if pid <= 0 {
		return false
	}
	if startHex != "" {
		proc, ok := defaultLookupPID(pid)
		if !ok || proc.StartTime.IsZero() {
			return false
		}
		return strconv.FormatInt(proc.StartTime.UnixNano(), 16) == startHex
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	return err == nil || errors.Is(err, syscall.EPERM)
}

func lookupProcfsPID(pid int) (Process, bool) {
	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat"))
	if err != nil {
		return Process{PID: pid}, false
	}

	name, fields, ok := parseProcStat(string(data))
	if !ok || len(fields) <= 19 {
		return Process{PID: pid}, false
	}

	ppid, err := strconv.Atoi(fields[1])
	if err != nil {
		return Process{PID: pid}, false
	}

	startTicks, _ := strconv.ParseInt(fields[19], 10, 64)
	return Process{
		PID:       pid,
		ParentPID: ppid,
		Name:      name,
		StartTime: startMarkerTime(startTicks),
	}, true
}

func parseProcStat(data string) (string, []string, bool) {
	open := strings.Index(data, "(")
	close := strings.LastIndex(data, ")")
	if open < 0 || close <= open {
		return "", nil, false
	}

	name := data[open+1 : close]
	fields := strings.Fields(data[close+1:])
	return name, fields, true
}

func startMarkerTime(marker int64) time.Time {
	if marker <= 0 {
		return time.Time{}
	}
	return time.Unix(0, marker)
}

func lookupPSPID(pid int) (Process, bool) {
	ppidText, ok := psField(pid, "ppid=")
	if !ok {
		return Process{PID: pid}, false
	}
	ppid, err := strconv.Atoi(strings.TrimSpace(ppidText))
	if err != nil {
		return Process{PID: pid}, false
	}

	name, _ := psField(pid, "comm=")
	startText, _ := psField(pid, "lstart=")
	startTime := parsePSStartTime(startText)

	return Process{
		PID:       pid,
		ParentPID: ppid,
		Name:      strings.TrimSpace(name),
		StartTime: startTime,
	}, true
}

func psField(pid int, format string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), psLookupTimeout)
	defer cancel()

	out, err := exec.CommandContext(ctx, "ps", "-p", strconv.Itoa(pid), "-o", format).Output()
	if err != nil {
		return "", false
	}
	value := strings.TrimSpace(string(out))
	if value == "" {
		return "", false
	}
	return value, true
}

func parsePSStartTime(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}

	layouts := []string{
		"Mon Jan 2 15:04:05 2006",
		"Mon Jan _2 15:04:05 2006",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return t
		}
	}
	return time.Time{}
}
