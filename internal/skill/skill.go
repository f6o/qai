package skill

import (
	_ "embed"
	"strings"
)

//go:embed SKILL.md
var markdown string

// Body returns the embedded SKILL.md without the YAML front matter.
func Body() string {
	s := markdown
	if !strings.HasPrefix(s, "---\n") {
		return s
	}
	rest := s[len("---\n"):]
	i := strings.Index(rest, "\n---\n")
	if i < 0 {
		return s
	}
	return strings.TrimLeft(rest[i+len("\n---\n"):], "\n")
}
