package accounts

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func tmpStore(t *testing.T) *Store {
	t.Helper()
	dir, err := os.MkdirTemp("", "docker-use-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return &Store{Root: dir}
}

func TestValidateName(t *testing.T) {
	bad := []string{"", ".", "..", "foo/bar", "foo\\bar", "   "}
	for _, name := range bad {
		if validateName(name) == nil {
			t.Errorf("expected error for %q", name)
		}
	}
	good := []string{"default", "work", "personal", "foo_bar", "foo-bar"}
	for _, name := range good {
		if err := validateName(name); err != nil {
			t.Errorf("unexpected error for %q: %v", name, err)
		}
	}
}

func TestList(t *testing.T) {
	s := tmpStore(t)
	for _, name := range []string{"beta", "alpha", "gamma"} {
		if err := os.MkdirAll(filepath.Join(s.Root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// create a file, which should be ignored
	if err := os.WriteFile(filepath.Join(s.Root, "notadir"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	accs, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(accs) != 3 {
		t.Fatalf("expected 3 accounts, got %d", len(accs))
	}
	want := []string{"alpha", "beta", "gamma"}
	for i, a := range accs {
		if a.Name != want[i] {
			t.Errorf("account[%d].Name = %q, want %q", i, a.Name, want[i])
		}
	}
}

func TestExistsAndPath(t *testing.T) {
	s := tmpStore(t)
	if s.Exists("missing") {
		t.Error("expected missing account to not exist")
	}
	p, err := s.Path("foo")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(p, filepath.Join(s.Root, "foo")) {
		t.Errorf("unexpected path: %s", p)
	}
}

func TestExport(t *testing.T) {
	s := tmpStore(t)
	if _, err := s.Export("missing"); err == nil {
		t.Error("expected error for missing account")
	}
	if err := os.MkdirAll(filepath.Join(s.Root, "foo"), 0o755); err != nil {
		t.Fatal(err)
	}
	line, err := s.Export("foo")
	if err != nil {
		t.Fatal(err)
	}
	want := `export DOCKER_CONFIG="` + filepath.Join(s.Root, "foo") + `"`
	if line != want {
		t.Errorf("export = %q, want %q", line, want)
	}
}

func TestRemove(t *testing.T) {
	s := tmpStore(t)
	if err := os.MkdirAll(filepath.Join(s.Root, "foo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := s.Remove("foo", false); err == nil {
		// stdin empty => abort
	} else if !strings.Contains(err.Error(), "aborted") {
		t.Fatalf("unexpected error: %v", err)
	}
	// still exists after abort
	if !s.Exists("foo") {
		t.Error("expected foo to still exist after aborted remove")
	}
	if err := s.Remove("foo", true); err != nil {
		t.Fatal(err)
	}
	if s.Exists("foo") {
		t.Error("expected foo to be removed")
	}
}

func TestStripCredentialHelpers(t *testing.T) {
	dir, err := os.MkdirTemp("", "docker-use-config-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	configPath := filepath.Join(dir, "config.json")
	input := `{
  "auths": {
    "https://index.docker.io/v1/": {}
  },
  "credsStore": "osxkeychain",
  "credHelpers": {
    "ghcr.io": "ghcr"
  }
}`
	if err := os.WriteFile(configPath, []byte(input), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := stripCredentialHelpers(configPath); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)
	if strings.Contains(out, "credsStore") {
		t.Error("expected credsStore to be removed")
	}
	if strings.Contains(out, "credHelpers") {
		t.Error("expected credHelpers to be removed")
	}
	if !strings.Contains(out, "auths") {
		t.Error("expected auths to be preserved")
	}
}

func TestDockerHubUsername(t *testing.T) {
	dir, err := os.MkdirTemp("", "docker-use-config-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	configPath := filepath.Join(dir, "config.json")
	auth := base64.StdEncoding.EncodeToString([]byte("alice:secret"))
	input := `{
  "auths": {
    "https://index.docker.io/v1/": {
      "auth": "` + auth + `"
    }
  }
}`
	if err := os.WriteFile(configPath, []byte(input), 0o600); err != nil {
		t.Fatal(err)
	}
	user, err := DockerHubUsername(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if user != "alice" {
		t.Errorf("user = %q, want alice", user)
	}
}
