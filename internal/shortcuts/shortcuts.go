package shortcuts

import "strings"

var notes = map[string]string{
	"clau-p":      "Claude planejando...",
	"clau-r":      "Claude revisando...",
	"clau-i":      "Claude implementando...",
	"clau-vp":     "Claude validando plano do Codex...",
	"clau-cp":     "Claude corrigindo plano com melhorias do Codex...",
	"dex-p":       "Codex planejando...",
	"dex-r":       "Codex revisando...",
	"dex-i":       "Codex implementando...",
	"dex-vp":      "Codex validando plano do Claude...",
	"dex-cp":      "Codex corrigindo plano com melhorias do Claude...",
	"clau-to-dex": "Claude gerando prompt para Codex...",
	"dex-to-clau": "Codex gerando prompt para Claude...",
}

func ResolveNote(text string) (string, bool) {
	resolved, ok := notes[strings.TrimSpace(text)]
	return resolved, ok
}
