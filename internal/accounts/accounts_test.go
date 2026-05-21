package accounts

import (
	"encoding/base64"
	"encoding/json"
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
	bad := []string{"", ".", "..", "foo/bar", "foo\\bar", "   ", "-x", "foo bar", "$x", "`x`", "semi;colon", "add", "list", "help", "init", strings.Repeat("a", 65)}
	for _, name := range bad {
		if validateName(name) == nil {
			t.Errorf("expected error for %q", name)
		}
	}
	good := []string{"default", "work", "personal", "foo_bar", "foo-bar", "foo.bar", strings.Repeat("a", 64)}
	for _, name := range good {
		if err := validateName(name); err != nil {
			t.Errorf("unexpected error for %q: %v", name, err)
		}
	}
}

func TestValidateUsername(t *testing.T) {
	bad := []string{"", "-alice", "foo bar", "$x", "`x`", "semi;colon", "pipe|x", "line\nbreak"}
	for _, username := range bad {
		if validateUsername(username) == nil {
			t.Errorf("expected error for %q", username)
		}
	}
	for _, username := range []string{"alice", "alice_123", "alice.name"} {
		if err := validateUsername(username); err != nil {
			t.Errorf("unexpected error for %q: %v", username, err)
		}
	}
}

func TestNewStorePermissions(t *testing.T) {
	root := filepath.Join(t.TempDir(), "accounts")
	t.Setenv("DOCKER_USE_DIR", root)
	if _, err := NewStore(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o700 {
		t.Fatalf("root mode = %o, want 700", got)
	}
}

func TestOpenStoreDoesNotCreateRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "accounts")
	t.Setenv("DOCKER_USE_DIR", root)
	store, err := OpenStore()
	if err != nil {
		t.Fatal(err)
	}
	if store.Root != root {
		t.Fatalf("root = %q, want %q", store.Root, root)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("OpenStore should not create root, stat err: %v", err)
	}
}

func TestList(t *testing.T) {
	s := tmpStore(t)
	for _, name := range []string{"beta", "alpha", "gamma"} {
		if err := os.MkdirAll(filepath.Join(s.Root, name), 0o700); err != nil {
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

func TestSymlinkAccountIsRejected(t *testing.T) {
	s := tmpStore(t)
	target := filepath.Join(t.TempDir(), "target")
	if err := os.MkdirAll(target, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(s.Root, "linked")); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}
	if s.Exists("linked") {
		t.Fatal("expected symlink account to not exist")
	}
	if _, err := s.Export("linked"); err == nil {
		t.Fatal("expected export to reject symlink account")
	}
	if err := s.Remove("linked"); err == nil {
		t.Fatal("expected remove to reject symlink account")
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("symlink target should not be removed: %v", err)
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
	want := filepath.Join(s.Root, "foo")
	if line != want {
		t.Errorf("export = %q, want %q", line, want)
	}
}

func TestRemove(t *testing.T) {
	s := tmpStore(t)
	if err := os.MkdirAll(filepath.Join(s.Root, "foo"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := s.Remove("foo"); err != nil {
		t.Fatal(err)
	}
	if s.Exists("foo") {
		t.Error("expected foo to be removed")
	}
}

func TestAddCreatesPrivateAccountAndForceReplaces(t *testing.T) {
	s := tmpStore(t)
	bin := t.TempDir()
	docker := filepath.Join(bin, "docker")
	script := `#!/bin/sh
config=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "--config" ]; then
    shift
    config="$1"
  fi
  shift
done
mkdir -p "$config"
cat > "$config/config.json" <<'JSON'
{"auths":{"https://index.docker.io/v1/":{"auth":"YWxpY2U6c2VjcmV0"}},"credsStore":"osxkeychain"}
JSON
`
	if err := os.WriteFile(docker, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	if err := s.Add(t.Context(), "foo", "alice", false); err != nil {
		t.Fatal(err)
	}
	accountPath := filepath.Join(s.Root, "foo")
	info, err := os.Stat(accountPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o700 {
		t.Fatalf("account mode = %o, want 700", got)
	}
	configData, err := os.ReadFile(filepath.Join(accountPath, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(configData), "credsStore") {
		t.Fatal("expected Add to strip credsStore")
	}
	username, err := s.Username("foo")
	if err != nil {
		t.Fatal(err)
	}
	if username != "alice" {
		t.Fatalf("username = %q, want alice", username)
	}
	if err := os.WriteFile(filepath.Join(accountPath, "marker"), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := s.Add(t.Context(), "foo", "alice", false); err == nil {
		t.Fatal("expected existing account error without force")
	}
	if err := s.Add(t.Context(), "foo", "alice", true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(accountPath, "marker")); !os.IsNotExist(err) {
		t.Fatalf("force should recreate account, marker stat err: %v", err)
	}
}

func TestAddCleansUpNewAccountOnDockerLoginFailure(t *testing.T) {
	s := tmpStore(t)
	bin := t.TempDir()
	docker := filepath.Join(bin, "docker")
	if err := os.WriteFile(docker, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	if err := s.Add(t.Context(), "foo", "alice", false); err == nil {
		t.Fatal("expected docker login failure")
	}
	if _, err := os.Stat(filepath.Join(s.Root, "foo")); !os.IsNotExist(err) {
		t.Fatalf("new account should be cleaned up, stat err: %v", err)
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
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("config mode = %o, want 600", got)
	}
}

func TestCurrentFromEnv(t *testing.T) {
	s := tmpStore(t)
	if err := os.MkdirAll(filepath.Join(s.Root, "foo"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DOCKER_CONFIG", filepath.Join(s.Root, "foo"))
	name, err := CurrentFromEnv(s)
	if err != nil {
		t.Fatal(err)
	}
	if name != "foo" {
		t.Fatalf("name = %q, want foo", name)
	}

	t.Setenv("DOCKER_CONFIG", filepath.Join(t.TempDir(), "external"))
	name, err = CurrentFromEnv(s)
	if err != nil {
		t.Fatal(err)
	}
	if name != "" {
		t.Fatalf("external name = %q, want empty", name)
	}
}

func TestSaveAndLoadCurrent(t *testing.T) {
	s := tmpStore(t)
	accountPath := filepath.Join(s.Root, "foo")
	if err := os.MkdirAll(accountPath, 0o700); err != nil {
		t.Fatal(err)
	}
	path, err := s.SaveCurrent("foo")
	if err != nil {
		t.Fatal(err)
	}
	if path != accountPath {
		t.Fatalf("path = %q, want %q", path, accountPath)
	}
	name, currentPath, err := s.Current()
	if err != nil {
		t.Fatal(err)
	}
	if name != "foo" || currentPath != accountPath {
		t.Fatalf("current = %q, %q; want foo, %q", name, currentPath, accountPath)
	}
	info, err := os.Stat(filepath.Join(s.Root, currentAccountFile))
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("current file mode = %o, want 600", got)
	}
}

func TestCurrentIgnoresStaleAccount(t *testing.T) {
	s := tmpStore(t)
	if err := os.WriteFile(filepath.Join(s.Root, currentAccountFile), []byte("missing\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	name, path, err := s.Current()
	if err != nil {
		t.Fatal(err)
	}
	if name != "" || path != "" {
		t.Fatalf("current = %q, %q; want empty", name, path)
	}
}

func TestUsernameMissingReturnsEmpty(t *testing.T) {
	s := tmpStore(t)
	if err := os.MkdirAll(filepath.Join(s.Root, "foo"), 0o700); err != nil {
		t.Fatal(err)
	}
	username, err := s.Username("foo")
	if err != nil {
		t.Fatal(err)
	}
	if username != "" {
		t.Fatalf("username = %q, want empty", username)
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

func TestDockerHubUsernameIgnoresContainsMatch(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	auth := base64.StdEncoding.EncodeToString([]byte("mallory:secret"))
	input := `{
  "auths": {
    "evil-docker.io.example.com": {
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
	if user != "" {
		t.Errorf("user = %q, want empty", user)
	}
}
