package shortcuts

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveNote(t *testing.T) {
	tests := map[string]string{
		"clau-p":      "Claude planejando...",
		"clau-r":      "Claude revisando...",
		"clau-i":      "Claude implementando...",
		"dex-p":       "Codex planejando...",
		"dex-r":       "Codex revisando...",
		"dex-i":       "Codex implementando...",
		"clau-to-dex": "Claude gerando prompt para Codex...",
		"dex-to-clau": "Codex gerando prompt para Claude...",
	}

	for input, expected := range tests {
		actual, ok := ResolveNote(input)
		require.True(t, ok, input)
		require.Equal(t, expected, actual)
	}
}

func TestResolveNoteUnknown(t *testing.T) {
	actual, ok := ResolveNote("texto livre")

	require.False(t, ok)
	require.Empty(t, actual)
}
