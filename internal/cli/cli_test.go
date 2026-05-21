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

func TestBareAccountPrintsHumanMessage(t *testing.T) {
	root := t.TempDir()
	t.Setenv("DOCKER_USE_DIR", root)
	t.Setenv("SHELL", "/opt/homebrew/bin/fish")
	accountPath := filepath.Join(root, "foo")
	if err := os.MkdirAll(accountPath, 0o700); err != nil {
		t.Fatal(err)
	}
	out, _, err := executeCommand(t, "foo")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `Account "foo" is available at `+accountPath) {
		t.Fatalf("output = %q, want account message with path %q", out, accountPath)
	}
	if !strings.Contains(out, "To switch this shell") {
		t.Fatalf("output = %q, want shell wrapper instruction", out)
	}
	if !strings.Contains(out, " init fish)") {
		t.Fatalf("output = %q, want init command instruction", out)
	}
}

func TestNoArgsPrintsShortMessage(t *testing.T) {
	out, _, err := executeCommand(t)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "No account selected") {
		t.Fatalf("output = %q, want no-account message", out)
	}
	if strings.Contains(out, "Available Commands") {
		t.Fatalf("output = %q, did not want full usage banner", out)
	}
}

func TestVersionFlag(t *testing.T) {
	out, _, err := executeCommand(t, "--version")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "docker-use version") {
		t.Fatalf("output = %q, want version output", out)
	}
}

func TestHiddenPathCommandPrintsOnlyPath(t *testing.T) {
	root := t.TempDir()
	t.Setenv("DOCKER_USE_DIR", root)
	accountPath := filepath.Join(root, "foo")
	if err := os.MkdirAll(accountPath, 0o700); err != nil {
		t.Fatal(err)
	}
	out, _, err := executeCommand(t, "__path", "foo")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != accountPath {
		t.Fatalf("output = %q, want path %q", out, accountPath)
	}
}

func TestHiddenSwitchPersistsCurrentAccount(t *testing.T) {
	root := t.TempDir()
	t.Setenv("DOCKER_USE_DIR", root)
	accountPath := filepath.Join(root, "foo")
	if err := os.MkdirAll(accountPath, 0o700); err != nil {
		t.Fatal(err)
	}
	out, _, err := executeCommand(t, "__switch", "foo")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != accountPath {
		t.Fatalf("switch output = %q, want path %q", out, accountPath)
	}
	out, _, err = executeCommand(t, "__current")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != accountPath {
		t.Fatalf("current output = %q, want path %q", out, accountPath)
	}
}

func TestUseCommandIsRemoved(t *testing.T) {
	_, _, err := executeCommand(t, "use", "foo")
	if err == nil {
		t.Fatal("expected error for removed use command")
	}
	if cmd, _, findErr := NewRootCommand().Find([]string{"use"}); findErr == nil && cmd.Name() == "use" {
		t.Fatal("use command should not be registered")
	}
}

func TestReservedAccountNameIsNotSwitchable(t *testing.T) {
	root := t.TempDir()
	t.Setenv("DOCKER_USE_DIR", root)
	if err := os.MkdirAll(filepath.Join(root, "list"), 0o700); err != nil {
		t.Fatal(err)
	}
	_, _, err := executeCommand(t, "__switch", "--", "list")
	if err == nil || !strings.Contains(err.Error(), "reserved") {
		t.Fatalf("expected reserved account error, got %v", err)
	}
}

func TestHiddenSwitchRejectsFlagLikeAccountNames(t *testing.T) {
	root := t.TempDir()
	t.Setenv("DOCKER_USE_DIR", root)
	_, _, err := executeCommand(t, "__switch", "--", "--version")
	if err == nil || !strings.Contains(err.Error(), "invalid account name") {
		t.Fatalf("expected invalid account error, got %v", err)
	}
}

func TestCurrentCommandPrintsPersistedAccount(t *testing.T) {
	root := t.TempDir()
	t.Setenv("DOCKER_USE_DIR", root)
	accountPath := filepath.Join(root, "foo")
	if err := os.MkdirAll(accountPath, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".current"), []byte("foo\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, err := executeCommand(t, "__current")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != accountPath {
		t.Fatalf("current output = %q, want %q", out, accountPath)
	}
}

func TestShellNameDoesNotGuessUnsupportedShell(t *testing.T) {
	t.Setenv("SHELL", "/usr/bin/nu")
	if got := shellName(); got != "" {
		t.Fatalf("shellName = %q, want empty", got)
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

func TestWhoamiFallsBackToStoredUsername(t *testing.T) {
	root := t.TempDir()
	t.Setenv("DOCKER_USE_DIR", root)
	accountPath := filepath.Join(root, "foo")
	if err := os.MkdirAll(accountPath, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(accountPath, "config.json"), []byte(`{"auths":{"https://index.docker.io/v1/":{}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(accountPath, ".username"), []byte("alice\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DOCKER_CONFIG", accountPath)
	out, _, err := executeCommand(t, "whoami")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Docker Hub user: alice") {
		t.Fatalf("output = %q, want stored username", out)
	}
}
