package config

import (
	"fmt"
	"os"
	"path/filepath"

	toml "github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// var (not const) so it can be overridden at build time via -ldflags -X.
var defaultAPIBase = "http://localhost:8601"

type Config struct {
	APIBase string  `koanf:"api_base"`
	Context Context `koanf:"context"`
}

type Context struct {
	WorkspaceID string `koanf:"workspace_id"`
	ProjectID   string `koanf:"project_id"`
}

// Path returns the config file location, honoring PADDI_CONFIG.
func Path() (string, error) {
	if p := os.Getenv("PADDI_CONFIG"); p != "" {
		return p, nil
	}
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "paddi", "config.toml"), nil
}

// Load returns the effective configuration: defaults, overridden by the
// config file, overridden by PADDI_* environment variables.
func Load() (*Config, error) {
	cfg := &Config{APIBase: defaultAPIBase}

	path, err := Path()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); err == nil {
		k := koanf.New(".")
		if err := k.Load(file.Provider(path), toml.Parser()); err != nil {
			return nil, fmt.Errorf("load config %s: %w", path, err)
		}
		if err := k.Unmarshal("", cfg); err != nil {
			return nil, fmt.Errorf("parse config %s: %w", path, err)
		}
		if cfg.APIBase == "" {
			cfg.APIBase = defaultAPIBase
		}
	}

	if v := os.Getenv("PADDI_API_BASE"); v != "" {
		cfg.APIBase = v
	}
	if v := os.Getenv("PADDI_PROJECT"); v != "" {
		cfg.Context.ProjectID = v
	}
	return cfg, nil
}

// Clear removes the config file. A missing file is not an error.
func Clear() error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Set updates a single key in the config file, creating the file if needed.
func Set(key, value string) error {
	path, err := Path()
	if err != nil {
		return err
	}
	k := koanf.New(".")
	if _, err := os.Stat(path); err == nil {
		if err := k.Load(file.Provider(path), toml.Parser()); err != nil {
			return fmt.Errorf("load config %s: %w", path, err)
		}
	}
	if err := k.Set(key, value); err != nil {
		return err
	}
	b, err := k.Marshal(toml.Parser())
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}
