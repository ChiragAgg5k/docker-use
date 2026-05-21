package accounts

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Store manages docker-use accounts on disk.
type Store struct {
	Root string
}

// Account represents a single docker-use account.
type Account struct {
	Name string
	Path string
}

// NewStore creates a Store, defaulting to ~/.docker-accounts or DOCKER_USE_DIR.
func NewStore() (*Store, error) {
	root := os.Getenv("DOCKER_USE_DIR")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		root = filepath.Join(home, ".docker-accounts")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	return &Store{Root: root}, nil
}

// validateName rejects path-unsafe account names.
func validateName(name string) error {
	if name == "" || name == "." || name == ".." {
		return fmt.Errorf("invalid account name: %q", name)
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("account name cannot contain path separators: %q", name)
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("account name cannot be whitespace only")
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
	return accs, nil
}

// Path returns the full directory path for an account.
func (s Store) Path(name string) (string, error) {
	if err := validateName(name); err != nil {
		return "", err
	}
	return filepath.Join(s.Root, name), nil
}

// Exists reports whether an account directory exists.
func (s Store) Exists(name string) bool {
	p, err := s.Path(name)
	if err != nil {
		return false
	}
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

// Export returns the shell export line for an account.
func (s Store) Export(name string) (string, error) {
	p, err := s.Path(name)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(p)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("account %q does not exist", name)
	}
	return fmt.Sprintf("export DOCKER_CONFIG=%q", p), nil
}

// Remove deletes an account directory with confirmation unless force is true.
func (s Store) Remove(name string, force bool) error {
	if !s.Exists(name) {
		return fmt.Errorf("account %q does not exist", name)
	}
	if !force {
		// We'll read from stdin; callers can pipe "y\n" in tests.
		fmt.Fprintf(os.Stderr, "Remove account %q? [y/N]: ", name)
		var answer string
		if _, err := fmt.Fscanln(os.Stdin, &answer); err != nil {
			return fmt.Errorf("aborted")
		}
		if strings.ToLower(strings.TrimSpace(answer)) != "y" {
			return fmt.Errorf("aborted")
		}
	}
	p, _ := s.Path(name)
	return os.RemoveAll(p)
}

// Add creates an account, runs docker login, and strips credential helpers.
func (s Store) Add(ctx context.Context, name, username string) error {
	if err := validateName(name); err != nil {
		return err
	}
	p := filepath.Join(s.Root, name)
	if err := os.MkdirAll(p, 0o755); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "docker", "--config", p, "login", "-u", username)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker login failed: %w", err)
	}

	configPath := filepath.Join(p, "config.json")
	if err := stripCredentialHelpers(configPath); err != nil {
		return fmt.Errorf("failed to strip credsStore: %w", err)
	}
	return nil
}

// stripCredentialHelpers removes credsStore and credHelpers from a Docker config file.
func stripCredentialHelpers(configPath string) error {
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
	out, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	return os.WriteFile(configPath, out, 0o600)
}

// CurrentFromEnv returns the account name inferred from DOCKER_CONFIG, or empty.
func CurrentFromEnv() string {
	p := os.Getenv("DOCKER_CONFIG")
	if p == "" {
		return ""
	}
	// Expects ~/.docker-accounts/<name>/config.json or at least ~/.docker-accounts/<name>
	// But DOCKER_CONFIG points to the *dir* containing config.json.
	base := filepath.Base(p)
	parent := filepath.Dir(p)
	if base == "config.json" {
		parent = filepath.Dir(parent)
		base = filepath.Base(parent)
	}
	return base
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
	for reg, cred := range config.Auths {
		if strings.Contains(reg, "docker.io") {
			if cred.Auth != "" {
				decoded, err := base64.StdEncoding.DecodeString(cred.Auth)
				if err == nil {
					parts := strings.SplitN(string(decoded), ":", 2)
					if len(parts) > 0 && parts[0] != "" {
						return parts[0], nil
					}
				}
			}
		}
	}
	return "", nil
}
