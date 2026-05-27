package shortcuts

import "strings"

var notes = map[string]string{
	"clau-p":      "Claude planejando...",
	"clau-r":      "Claude revisando...",
	"clau-i":      "Claude implementando...",
	"dex-p":       "Codex planejando...",
	"dex-r":       "Codex revisando...",
	"dex-i":       "Codex implementando...",
	"clau-to-dex": "Claude gerando prompt para Codex...",
	"dex-to-clau": "Codex gerando prompt para Claude...",
}

func ResolveNote(text string) (string, bool) {
	resolved, ok := notes[strings.TrimSpace(text)]
	return resolved, ok
}
