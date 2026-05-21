package shell

import (
	"strings"
	"testing"
)

func TestInitScript(t *testing.T) {
	for _, shell := range []string{"zsh", "bash", "fish"} {
		script, err := InitScript(shell)
		if err != nil {
			t.Fatalf("InitScript(%q) error: %v", shell, err)
		}
		if script == "" {
			t.Fatalf("InitScript(%q) returned empty script", shell)
		}
		// Ensure the template was filled with the binary path.
		if !strings.Contains(script, "docker-use") {
			t.Errorf("%s script missing binary path reference", shell)
		}
		if strings.Contains(script, "eval") {
			t.Errorf("%s script must not use eval", shell)
		}
		if !strings.Contains(script, "DOCKER_CONFIG") {
			t.Errorf("%s script missing DOCKER_CONFIG assignment", shell)
		}
	}
}

func TestInitScriptUnsupported(t *testing.T) {
	if _, err := InitScript("pwsh"); err == nil {
		t.Error("expected error for unsupported shell")
	}
}
