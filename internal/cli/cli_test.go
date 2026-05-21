package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func executeCommand(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := NewRootCommand()
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errOut.String(), err
}

func TestUsePrintsOnlyPath(t *testing.T) {
	root := t.TempDir()
	t.Setenv("DOCKER_USE_DIR", root)
	accountPath := filepath.Join(root, "foo")
	if err := os.MkdirAll(accountPath, 0o700); err != nil {
		t.Fatal(err)
	}
	out, _, err := executeCommand(t, "use", "foo")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != accountPath {
		t.Fatalf("output = %q, want path %q", out, accountPath)
	}
	if strings.Contains(out, "export") {
		t.Fatalf("use output must not contain shell code: %q", out)
	}
}

func TestInitRejectsInvalidShell(t *testing.T) {
	_, _, err := executeCommand(t, "init", "pwsh")
	if err == nil {
		t.Fatal("expected invalid shell error")
	}
}

func TestRemoveConfirmationUsesCommandInput(t *testing.T) {
	root := t.TempDir()
	t.Setenv("DOCKER_USE_DIR", root)
	accountPath := filepath.Join(root, "foo")
	if err := os.MkdirAll(accountPath, 0o700); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"remove", "foo"})
	cmd.SetIn(strings.NewReader("n\n"))
	var errOut bytes.Buffer
	cmd.SetErr(&errOut)
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "aborted") {
		t.Fatalf("expected aborted error, got %v", err)
	}
	if _, err := os.Stat(accountPath); err != nil {
		t.Fatalf("account should remain after no: %v", err)
	}

	cmd = NewRootCommand()
	cmd.SetArgs([]string{"remove", "foo"})
	cmd.SetIn(strings.NewReader("y\n"))
	cmd.SetErr(&errOut)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(accountPath); !os.IsNotExist(err) {
		t.Fatalf("account should be removed, stat err: %v", err)
	}
}

func TestWhoamiCorruptConfigReturnsError(t *testing.T) {
	root := t.TempDir()
	t.Setenv("DOCKER_USE_DIR", root)
	accountPath := filepath.Join(root, "foo")
	if err := os.MkdirAll(accountPath, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(accountPath, "config.json"), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DOCKER_CONFIG", accountPath)
	_, _, err := executeCommand(t, "whoami")
	if err == nil {
		t.Fatal("expected corrupt config error")
	}
}
