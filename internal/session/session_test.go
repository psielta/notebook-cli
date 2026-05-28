package session

import (
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestResolveEnvVarTakesPrecedence(t *testing.T) {
	r := &Resolver{
		Env:       func(string) string { return "my-session" },
		PPID:      func() int { return 4321 },
		LookupPID: nil,
	}
	require.Equal(t, "my-session", r.Resolve())
}

func TestResolveExplicitEnvVarBeatsTerminalEnv(t *testing.T) {
	r := &Resolver{
		Env: func(name string) string {
			switch name {
			case EnvVar:
				return "my-session"
			case "WT_SESSION":
				return "terminal-session"
			default:
				return ""
			}
		},
		PPID: func() int { return 4321 },
	}

	require.Equal(t, "my-session", r.Resolve())
}

func TestResolveEnvVarRejectsInvalidPattern(t *testing.T) {
	cases := []string{
		"../escape",
		"name with spaces",
		"slash/inside",
		"sh-123",
		"SH-123",
		strings.Repeat("a", 129),
	}
	for _, bad := range cases {
		t.Run(bad, func(t *testing.T) {
			r := &Resolver{
				Env: func(name string) string {
					if name == EnvVar {
						return bad
					}
					return ""
				},
				PPID: func() int { return 0 },
			}
			require.Equal(t, DefaultSessionID, r.Resolve())
		})
	}
}

func TestResolveTerminalEnvBeforeProcessTree(t *testing.T) {
	r := &Resolver{
		Env: func(name string) string {
			if name == "WT_SESSION" {
				return "abc-123"
			}
			return ""
		},
		PPID: func() int { return 100 },
		LookupPID: func(pid int) (Process, bool) {
			return Process{PID: pid, Name: "powershell.exe", StartTime: time.Unix(0, 0xABCDEF)}, true
		},
	}

	require.Equal(t, FormatDerivedSessionID("wt", "WT_SESSION=abc-123"), r.Resolve())
}

func TestResolveOSTerminalIDBeforeProcessTree(t *testing.T) {
	r := &Resolver{
		Env:        func(string) string { return "" },
		TerminalID: func() (string, bool) { return FormatDerivedSessionID("tty", "dev:pty"), true },
		PPID:       func() int { return 100 },
		LookupPID: func(pid int) (Process, bool) {
			return Process{PID: pid, Name: "bash", StartTime: time.Unix(0, 0xABCDEF)}, true
		},
	}

	require.Equal(t, FormatDerivedSessionID("tty", "dev:pty"), r.Resolve())
}

func TestResolveWalksToRecognizedShell(t *testing.T) {
	tree := map[int]Process{
		100: {PID: 100, ParentPID: 50, Name: "make.exe"},
		50:  {PID: 50, ParentPID: 10, Name: "node.exe"},
		10:  {PID: 10, ParentPID: 4, Name: "powershell.exe", StartTime: time.Unix(0, 0xABCDEF)},
		4:   {PID: 4, ParentPID: 0, Name: "System"},
	}
	r := &Resolver{
		Env:       func(string) string { return "" },
		PPID:      func() int { return 100 },
		LookupPID: func(pid int) (Process, bool) { p, ok := tree[pid]; return p, ok },
	}

	got := r.Resolve()
	require.Equal(t, "sh-10-"+strconv.FormatInt(0xABCDEF, 16), got)
}

func TestResolveFallbacksToPPIDWhenNoShellFound(t *testing.T) {
	tree := map[int]Process{
		7: {PID: 7, ParentPID: 3, Name: "go.exe", StartTime: time.Unix(0, 0xCAFE)},
		3: {PID: 3, ParentPID: 1, Name: "init"},
	}
	r := &Resolver{
		Env:       func(string) string { return "" },
		PPID:      func() int { return 7 },
		LookupPID: func(pid int) (Process, bool) { p, ok := tree[pid]; return p, ok },
	}

	got := r.Resolve()
	require.Equal(t, "sh-7-"+strconv.FormatInt(0xCAFE, 16), got)
}

func TestResolveFallbacksToPIDOnlyWhenLookupFails(t *testing.T) {
	r := &Resolver{
		Env:       func(string) string { return "" },
		PPID:      func() int { return 9999 },
		LookupPID: func(int) (Process, bool) { return Process{PID: 9999}, false },
	}

	require.Equal(t, "sh-9999", r.Resolve())
}

func TestResolveDefaultWhenPPIDUnavailable(t *testing.T) {
	r := &Resolver{
		Env:  func(string) string { return "" },
		PPID: func() int { return 1 },
	}
	require.Equal(t, DefaultSessionID, r.Resolve())
}

func TestResolveStopsOnCycle(t *testing.T) {
	tree := map[int]Process{
		20: {PID: 20, ParentPID: 30, Name: "a.exe"},
		30: {PID: 30, ParentPID: 20, Name: "b.exe"},
	}
	r := &Resolver{
		Env:       func(string) string { return "" },
		PPID:      func() int { return 20 },
		LookupPID: func(pid int) (Process, bool) { p, ok := tree[pid]; return p, ok },
	}

	got := r.Resolve()
	require.Equal(t, "sh-20", got)
}

func TestResolveRespectsDepthLimit(t *testing.T) {
	tree := map[int]Process{}
	for i := 100; i <= 200; i++ {
		tree[i] = Process{PID: i, ParentPID: i + 1, Name: "x.exe"}
	}
	r := &Resolver{
		Env:       func(string) string { return "" },
		PPID:      func() int { return 100 },
		LookupPID: func(pid int) (Process, bool) { p, ok := tree[pid]; return p, ok },
	}

	got := r.Resolve()
	require.Equal(t, "sh-100", got)
}

func TestFormatProcessIDStartTime(t *testing.T) {
	p := Process{PID: 42, StartTime: time.Unix(0, 0xDEADBEEF)}
	require.Equal(t, "sh-42-"+strconv.FormatInt(0xDEADBEEF, 16), FormatProcessID(p))
}

func TestFormatProcessIDWithoutStartTime(t *testing.T) {
	require.Equal(t, "sh-42", FormatProcessID(Process{PID: 42}))
}

func TestRecognizedShellAcceptsUnixNamesAndPaths(t *testing.T) {
	require.True(t, isRecognizedShell("/bin/bash"))
	require.True(t, isRecognizedShell("C:\\Program Files\\PowerShell\\7\\pwsh.exe"))
	require.True(t, isRecognizedShell("fish"))
	require.False(t, isRecognizedShell("node"))
}

func TestDefaultResolverReturnsNonEmpty(t *testing.T) {
	require.NotEmpty(t, DefaultResolver().Resolve())
}

func TestIsAliveForCurrentProcess(t *testing.T) {
	require.True(t, IsAlive(os.Getpid(), ""))
}

func TestIsAliveForNonExistentProcess(t *testing.T) {
	require.False(t, IsAlive(0x7FFFFFFE, ""))
}
