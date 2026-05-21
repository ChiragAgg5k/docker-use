package accounts

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

var accountNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,63}$`)

var reservedAccountNames = map[string]struct{}{
	"add":        {},
	"completion": {},
	"help":       {},
	"init":       {},
	"list":       {},
	"ls":         {},
	"remove":     {},
	"rm":         {},
	"whoami":     {},
}

// Store manages docker-use accounts on disk.
type Store struct {
	Root string
}

// Account represents a single docker-use account.
type Account struct {
	Name string
	Path string
}

const (
	currentAccountFile = ".current"
	usernameFile       = ".username"
)

// NewStore creates a Store, defaulting to ~/.docker-accounts or DOCKER_USE_DIR.
func NewStore() (*Store, error) {
	root, err := storeRoot()
	if err != nil {
		return nil, err
	}
	if info, err := os.Lstat(root); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("account root %q is a symlink", root)
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, err
	}
	if err := os.Chmod(root, 0o700); err != nil {
		return nil, err
	}
	return &Store{Root: root}, nil
}

// OpenStore creates a Store without creating or chmodding the root directory.
func OpenStore() (*Store, error) {
	root, err := storeRoot()
	if err != nil {
		return nil, err
	}
	return &Store{Root: root}, nil
}

func storeRoot() (string, error) {
	root := os.Getenv("DOCKER_USE_DIR")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		root = filepath.Join(home, ".docker-accounts")
	}
	return root, nil
}

// validateName rejects path-unsafe account names.
func validateName(name string) error {
	if !accountNamePattern.MatchString(name) {
		return fmt.Errorf("invalid account name: %q", name)
	}
	if _, ok := reservedAccountNames[name]; ok {
		return fmt.Errorf("account name %q is reserved", name)
	}
	return nil
}

func validateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if strings.HasPrefix(username, "-") {
		return fmt.Errorf("username cannot start with '-'")
	}
	for _, r := range username {
		if unicode.IsSpace(r) || unicode.IsControl(r) || strings.ContainsRune(";&|`$<>(){}[]*?!\\\"'", r) {
			return fmt.Errorf("username contains unsafe character %q", r)
		}
	}
	return nil
}

// List returns sorted account names.
func (s Store) List() ([]Account, error) {
	entries, err := os.ReadDir(s.Root)
	if err != nil {
		return nil, err
	}
	var accs []Account
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if validateName(name) != nil {
			continue
		}
		accs = append(accs, Account{
			Name: name,
			Path: filepath.Join(s.Root, name),
		})
	}
	sort.Slice(accs, func(i, j int) bool { return accs[i].Name < accs[j].Name })
	return accs, nil
}

// Path returns the full directory path for an account.
func (s Store) Path(name string) (string, error) {
	if err := validateName(name); err != nil {
		return "", err
	}
	return s.accountPath(name), nil
}

func (s Store) accountPath(name string) string {
	return filepath.Join(s.Root, name)
}

func (s Store) existingValidatedAccountPath(name string) (string, error) {
	p := s.accountPath(name)
	info, err := os.Lstat(p)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("account %q is a symlink", name)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("account %q is not a directory", name)
	}
	return p, nil
}

func (s Store) existingAccountPath(name string) (string, error) {
	if err := validateName(name); err != nil {
		return "", err
	}
	return s.existingValidatedAccountPath(name)
}

// Exists reports whether an account directory exists.
func (s Store) Exists(name string) bool {
	_, err := s.existingAccountPath(name)
	return err == nil
}

// Export returns the account directory for shell wrappers.
func (s Store) Export(name string) (string, error) {
	if err := validateName(name); err != nil {
		return "", err
	}
	p, err := s.existingValidatedAccountPath(name)
	if err != nil {
		return "", fmt.Errorf("account %q does not exist", name)
	}
	return p, nil
}

// SaveCurrent persists the account selected by the shell wrapper.
func (s Store) SaveCurrent(name string) (string, error) {
	p, err := s.Export(name)
	if err != nil {
		return "", err
	}
	if err := atomicWriteFile(filepath.Join(s.Root, currentAccountFile), []byte(name+"\n"), 0o600); err != nil {
		return "", err
	}
	return p, nil
}

// Current returns the persisted account and path, or empty values when unset.
func (s Store) Current() (string, string, error) {
	currentPath := filepath.Join(s.Root, currentAccountFile)
	info, err := os.Lstat(currentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", nil
		}
		return "", "", err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", "", fmt.Errorf("current account file is a symlink")
	}
	data, err := os.ReadFile(currentPath)
	if err != nil {
		return "", "", err
	}
	name := strings.TrimSpace(string(data))
	if name == "" {
		return "", "", nil
	}
	if err := validateName(name); err != nil {
		return "", "", err
	}
	p, err := s.existingValidatedAccountPath(name)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", nil
		}
		return "", "", err
	}
	return name, p, nil
}

// Username returns the Docker Hub username saved when the account was added.
func (s Store) Username(name string) (string, error) {
	p, err := s.Export(name)
	if err != nil {
		return "", err
	}
	usernamePath := filepath.Join(p, usernameFile)
	info, err := os.Lstat(usernamePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("username file for account %q is a symlink", name)
	}
	data, err := os.ReadFile(usernamePath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// Remove deletes an account directory.
func (s Store) Remove(name string) error {
	p, err := s.existingAccountPath(name)
	if err != nil {
		return fmt.Errorf("account %q does not exist", name)
	}
	root, err := filepath.Abs(s.Root)
	if err != nil {
		return err
	}
	target, err := filepath.Abs(p)
	if err != nil {
		return err
	}
	root = filepath.Clean(root)
	target = filepath.Clean(target)
	if target != root && !strings.HasPrefix(target, root+string(os.PathSeparator)) {
		return fmt.Errorf("refusing to remove path outside account root: %s", target)
	}
	return os.RemoveAll(p)
}

// Add creates an account, runs docker login, and strips credential helpers.
func (s Store) Add(ctx context.Context, name, username string, force bool) error {
	if err := validateName(name); err != nil {
		return err
	}
	if err := validateUsername(username); err != nil {
		return err
	}
	p := filepath.Join(s.Root, name)
	if _, err := s.existingAccountPath(name); err == nil {
		if !force {
			return fmt.Errorf("account %q already exists; use --force to replace it", name)
		}
		if err := s.Remove(name); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(p, 0o700); err != nil {
		return err
	}
	if err := os.Chmod(p, 0o700); err != nil {
		_ = os.RemoveAll(p)
		return err
	}

	cmd := exec.CommandContext(ctx, "docker", "--config", p, "login", "-u", username)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		_ = os.RemoveAll(p)
		return fmt.Errorf("docker login failed: %w", err)
	}

	configPath := filepath.Join(p, "config.json")
	if err := normalizeDockerConfig(configPath); err != nil {
		_ = os.RemoveAll(p)
		return fmt.Errorf("failed to normalize docker config: %w", err)
	}
	if err := os.WriteFile(filepath.Join(p, usernameFile), []byte(username+"\n"), 0o600); err != nil {
		_ = os.RemoveAll(p)
		return fmt.Errorf("failed to save username: %w", err)
	}
	return nil
}

// normalizeDockerConfig strips credential helpers and points the account at the
// user's default cli-plugins directory so `docker compose` and friends keep
// working when DOCKER_CONFIG is redirected to the account.
func normalizeDockerConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}
	delete(config, "credsStore")
	delete(config, "credHelpers")
	if home, err := os.UserHomeDir(); err == nil {
		config["cliPluginsExtraDirs"] = []string{filepath.Join(home, ".docker", "cli-plugins")}
	}
	out, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	return atomicWriteFile(configPath, out, 0o600)
}

func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	return fsyncDir(dir)
}

func fsyncDir(dir string) error {
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.Sync(); err != nil {
		return err
	}
	return nil
}

// CurrentFromEnv returns the current account when DOCKER_CONFIG points under the store root.
func CurrentFromEnv(store *Store) (string, error) {
	if store == nil {
		var err error
		store, err = NewStore()
		if err != nil {
			return "", err
		}
	}
	p := os.Getenv("DOCKER_CONFIG")
	if p == "" {
		return "", nil
	}
	root, err := filepath.Abs(store.Root)
	if err != nil {
		return "", err
	}
	config, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	root = filepath.Clean(root)
	config = filepath.Clean(config)
	if config == root || !strings.HasPrefix(config, root+string(os.PathSeparator)) {
		return "", nil
	}
	rel, err := filepath.Rel(root, config)
	if err != nil {
		return "", err
	}
	parts := strings.Split(rel, string(os.PathSeparator))
	if len(parts) != 1 {
		return "", nil
	}
	name := parts[0]
	if _, err := store.existingAccountPath(name); err != nil {
		return "", nil
	}
	return name, nil
}

// DockerHubUsername reads the auths section of a Docker config and returns the
// username for index.docker.io, if present.
func DockerHubUsername(configPath string) (string, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}
	var config struct {
		Auths map[string]struct {
			Auth string `json:"auth"`
		} `json:"auths"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return "", err
	}
	for _, reg := range []string{"https://index.docker.io/v1/", "index.docker.io", "registry-1.docker.io", "docker.io"} {
		cred, ok := config.Auths[reg]
		if !ok || cred.Auth == "" {
			continue
		}
		decoded, err := base64.StdEncoding.DecodeString(cred.Auth)
		if err == nil {
			parts := strings.SplitN(string(decoded), ":", 2)
			if len(parts) > 0 && parts[0] != "" {
				return parts[0], nil
			}
		}
	}
	return "", nil
}
