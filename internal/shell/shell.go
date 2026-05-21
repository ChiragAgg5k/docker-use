package shell

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
)

//go:embed zsh.sh
var zshTemplate string

//go:embed bash.sh
var bashTemplate string

//go:embed fish.fish
var fishTemplate string

// InitScript returns the shell integration script for the named shell.
func InitScript(shell string) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	var tmpl string
	switch shell {
	case "zsh":
		tmpl = zshTemplate
	case "bash":
		tmpl = bashTemplate
	case "fish":
		tmpl = fishTemplate
	default:
		return "", fmt.Errorf("unsupported shell: %q", shell)
	}
	return strings.ReplaceAll(tmpl, "{{ .Binary }}", exe), nil
}
