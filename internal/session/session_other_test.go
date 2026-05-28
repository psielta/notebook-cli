//go:build !windows

package session

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseProcStat(t *testing.T) {
	fields := []string{
		"S", "45", "1", "2", "3",
		"4", "5", "6", "7", "8",
		"9", "10", "11", "12", "13",
		"14", "15", "16", "17", "9999",
	}

	name, gotFields, ok := parseProcStat("123 (bash) " + strings.Join(fields, " "))

	require.True(t, ok)
	require.Equal(t, "bash", name)
	require.Equal(t, "45", gotFields[1])
	require.Equal(t, "9999", gotFields[19])
}

func TestParsePSStartTime(t *testing.T) {
	got := parsePSStartTime("Thu May 28 10:23:05 2026")

	require.Equal(t, time.Date(2026, 5, 28, 10, 23, 5, 0, time.Local), got)
}
