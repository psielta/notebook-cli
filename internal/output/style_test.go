package output

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStyleDisabledForNonTerminal(t *testing.T) {
	var out bytes.Buffer
	style := NewStyle(&out)

	require.False(t, style.Enabled())
	require.Equal(t, "ok", style.Success("ok"))
}

func TestStyleEnabledForTerminal(t *testing.T) {
	style := newStyle(true, false, "xterm")

	require.True(t, style.Enabled())
	require.Equal(t, "\x1b[32mok\x1b[0m", style.Success("ok"))
}

func TestStyleDisabledByNoColor(t *testing.T) {
	style := newStyle(true, true, "xterm")

	require.False(t, style.Enabled())
	require.Equal(t, "ok", style.Success("ok"))
}

func TestStyleDisabledForDumbTerminal(t *testing.T) {
	style := newStyle(true, false, "dumb")

	require.False(t, style.Enabled())
	require.Equal(t, "ok", style.Success("ok"))
}
