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
		if strings.Contains(script, "path=") || strings.Contains(script, "set -l path") {
			t.Errorf("%s script must not assign to path; zsh treats it as a special PATH-tied variable", shell)
		}
		if !strings.Contains(script, "docker_use_config") {
			t.Errorf("%s script missing bare account switch variable", shell)
		}
		if !strings.Contains(script, "__switch") {
			t.Errorf("%s script missing hidden switch command", shell)
		}
		if !strings.Contains(script, "__current") {
			t.Errorf("%s script missing current account restore", shell)
		}
		for _, command := range []string{"add", "remove", "whoami", "completion"} {
			if strings.Contains(script, `"`+command+`"`) || strings.Contains(script, " "+command+" ") {
				t.Errorf("%s script should not hardcode command allowlist entry %q", shell, command)
			}
		}
	}
}

func TestInitScriptUnsupported(t *testing.T) {
	if _, err := InitScript("pwsh"); err == nil {
		t.Error("expected error for unsupported shell")
	}
}
