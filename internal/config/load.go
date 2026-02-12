package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load reads and validates the config from ConfigFile().
func Load() (Config, error) {
	data, err := os.ReadFile(ConfigFile())
	if err != nil {
		return Config{}, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}

	if errs := Validate(cfg); len(errs) > 0 {
		return Config{}, fmt.Errorf("invalid config: %w", errs[0])
	}

	return cfg, nil
}

// Save atomically writes cfg to ConfigFile().
// It writes to a temp file then renames, preventing partial reads.
func Save(cfg Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	path := ConfigFile()
	tmp := path + ".tmp"

	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("writing temp config: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp) // best-effort cleanup
		return fmt.Errorf("renaming temp config: %w", err)
	}

	return nil
}

// LoadProjectConfig reads a per-project .hatch.yml file.
func LoadProjectConfig(path string) (ProjectConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ProjectConfig{}, fmt.Errorf("reading project config: %w", err)
	}

	var pc ProjectConfig
	if err := yaml.Unmarshal(data, &pc); err != nil {
		return ProjectConfig{}, fmt.Errorf("parsing project config: %w", err)
	}

	if pc.Domain == "" {
		return ProjectConfig{}, fmt.Errorf("project config: domain is required")
	}
	if len(pc.Services) == 0 {
		return ProjectConfig{}, fmt.Errorf("project config: at least one service is required")
	}

	return pc, nil
}

// LoadRaw reads the config file without validation, useful for merging.
func LoadRaw() (Config, error) {
	data, err := os.ReadFile(ConfigFile())
	if err != nil {
		return Config{}, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.Projects == nil {
		cfg.Projects = make(map[string]Project)
	}

	return cfg, nil
}

// EnsureConfigDir creates the config directory and subdirectories.
func EnsureConfigDir() error {
	dirs := []string{Dir(), CertsDir()}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", d, err)
		}
	}
	return nil
}

// EnsureConfigFile creates a default config file if one doesn't exist.
func EnsureConfigFile() error {
	path := ConfigFile()
	if _, err := os.Stat(path); err == nil {
		return nil // already exists
	}

	cfg := DefaultConfig()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling default config: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// Init ensures the config directory structure and default config file exist.
func Init() error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}
	return EnsureConfigFile()
}

// ConfigFileDir returns the directory containing the config file,
// used by the watcher to watch the directory instead of the file.
func ConfigFileDir() string {
	return filepath.Dir(ConfigFile())
}
