// Package session resolves a stable identifier for the terminal session
// hosting the current process. The identifier is used to scope per-session
// state (such as the "current notebook" selection) so that multiple
// PowerShell windows do not share global state.
package session

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// EnvVar is the optional environment variable that overrides session
	// identification. When set, valid, and not using the reserved SessionPrefix,
	// its value is used as the session id verbatim.
	EnvVar = "NB_SESSION_ID"

	// DefaultSessionID is the sentinel used when no parent process can be
	// inspected. Behaves like the previous global state.
	DefaultSessionID = "_default"

	// SessionPrefix marks ids derived from process inspection. Prune only
	// considers files whose name starts with this prefix.
	SessionPrefix = "sh-"

	terminalPrefix = "term-"

	walkDepthLimit = 10
)

var recognizedShells = map[string]struct{}{
	"bash":                {},
	"powershell.exe":      {},
	"powershell":          {},
	"pwsh.exe":            {},
	"pwsh":                {},
	"cmd.exe":             {},
	"cmd":                 {},
	"wt.exe":              {},
	"windowsterminal.exe": {},
	"bash.exe":            {},
	"sh.exe":              {},
	"sh":                  {},
	"zsh.exe":             {},
	"zsh":                 {},
	"fish.exe":            {},
	"fish":                {},
	"dash":                {},
	"ksh":                 {},
	"mksh":                {},
	"tcsh":                {},
	"csh":                 {},
	"nu.exe":              {},
	"nu":                  {},
	"xonsh":               {},
	"elvish":              {},
}

var terminalEnvVars = []struct {
	name   string
	prefix string
}{
	{name: "WT_SESSION", prefix: "wt"},
	{name: "TERM_SESSION_ID", prefix: "term"},
	{name: "ITERM_SESSION_ID", prefix: "iterm"},
	{name: "TMUX_PANE", prefix: "tmux"},
	{name: "STY", prefix: "screen"},
	{name: "SSH_TTY", prefix: "ssh"},
	{name: "WEZTERM_PANE", prefix: "wezterm"},
	{name: "KITTY_WINDOW_ID", prefix: "kitty"},
	{name: "ALACRITTY_WINDOW_ID", prefix: "alacritty"},
	{name: "CONEMUHWND", prefix: "conemu"},
	{name: "ConEmuPID", prefix: "conemu"},
}

var sessionIDPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,128}$`)

// Process is a snapshot of a process used for session resolution.
type Process struct {
	PID       int
	ParentPID int
	Name      string
	StartTime time.Time
}

// LookupFunc retrieves info about a process. The boolean is false when the
// process could not be located.
type LookupFunc func(pid int) (Process, bool)

// Resolver produces a stable session identifier. All dependencies are
// injectable to allow deterministic tests.
type Resolver struct {
	Env        func(string) string
	TerminalID func() (string, bool)
	PPID       func() int
	LookupPID  LookupFunc
}

// DefaultResolver returns a Resolver wired with OS defaults.
func DefaultResolver() *Resolver {
	return &Resolver{
		Env:        os.Getenv,
		TerminalID: defaultTerminalID,
		PPID:       os.Getppid,
		LookupPID:  defaultLookupPID,
	}
}

// Resolve returns the session identifier following the fallback chain:
//  1. NB_SESSION_ID env var (validated and checked against SessionPrefix).
//  2. Known terminal/session environment variables inherited through wrappers.
//  3. OS-specific terminal id, such as a Unix TTY or Windows console window.
//  4. Walk up the process tree (max walkDepthLimit hops) for a recognized
//     shell, formatted as "sh-<pid>-<creationHex>".
//  5. Direct parent PID (same format).
//  6. "sh-<ppid>" without creation time when LookupPID fails.
//  7. DefaultSessionID when even os.Getppid returns an unusable value.
func (r *Resolver) Resolve() string {
	if r.Env != nil {
		if v := strings.TrimSpace(r.Env(EnvVar)); isValidEnvSessionID(v) {
			return v
		}
	}

	if id, ok := r.terminalEnvSessionID(); ok {
		return id
	}

	if r.TerminalID != nil {
		if id, ok := r.TerminalID(); ok && isValidDerivedSessionID(id) {
			return id
		}
	}

	if r.PPID == nil {
		return DefaultSessionID
	}

	ppid := r.PPID()
	if ppid <= 1 {
		return DefaultSessionID
	}

	if id, ok := r.walkForShell(ppid); ok {
		return id
	}

	if r.LookupPID != nil {
		if proc, ok := r.LookupPID(ppid); ok {
			return FormatProcessID(proc)
		}
	}

	return SessionPrefix + strconv.Itoa(ppid)
}

func isValidEnvSessionID(v string) bool {
	if v == "" || !sessionIDPattern.MatchString(v) {
		return false
	}
	return !strings.HasPrefix(strings.ToLower(v), SessionPrefix)
}

func isValidDerivedSessionID(v string) bool {
	return v != "" && sessionIDPattern.MatchString(v)
}

func (r *Resolver) terminalEnvSessionID() (string, bool) {
	if r.Env == nil {
		return "", false
	}
	for _, candidate := range terminalEnvVars {
		value := strings.TrimSpace(r.Env(candidate.name))
		if value == "" {
			continue
		}
		return FormatDerivedSessionID(candidate.prefix, candidate.name+"="+value), true
	}
	return "", false
}

func (r *Resolver) walkForShell(start int) (string, bool) {
	if r.LookupPID == nil {
		return "", false
	}
	seen := map[int]bool{}
	current := start
	for depth := 0; depth < walkDepthLimit; depth++ {
		if current <= 1 || seen[current] {
			return "", false
		}
		seen[current] = true

		proc, ok := r.LookupPID(current)
		if !ok {
			return "", false
		}
		if isRecognizedShell(proc.Name) {
			return FormatProcessID(proc), true
		}
		if proc.ParentPID == proc.PID || proc.ParentPID <= 1 {
			return "", false
		}
		current = proc.ParentPID
	}
	return "", false
}

func isRecognizedShell(name string) bool {
	_, ok := recognizedShells[processBaseName(name)]
	return ok
}

func processBaseName(name string) string {
	base := strings.Trim(strings.ToLower(strings.TrimSpace(name)), "\x00")
	base = filepath.Base(base)
	if i := strings.LastIndex(base, `\`); i >= 0 {
		base = base[i+1:]
	}
	return base
}

// FormatDerivedSessionID returns a stable, filesystem-safe identifier for
// terminal signals inherited through process wrappers such as npm or npx.
func FormatDerivedSessionID(prefix, value string) string {
	sum := sha256.Sum256([]byte(value))
	return terminalPrefix + prefix + "-" + hex.EncodeToString(sum[:8])
}

// FormatProcessID encodes a Process as a session identifier. When StartTime
// is unknown, only the PID is encoded (less robust against PID reuse).
func FormatProcessID(p Process) string {
	if p.StartTime.IsZero() {
		return SessionPrefix + strconv.Itoa(p.PID)
	}
	return SessionPrefix + strconv.Itoa(p.PID) + "-" + strconv.FormatInt(p.StartTime.UnixNano(), 16)
}

// IsAlive reports whether the given PID is running and (when startHex is
// non-empty) has the expected creation time. Used by the prune routine to
// detect orphan session files.
func IsAlive(pid int, startHex string) bool {
	return defaultIsAlive(pid, startHex)
}
