package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the aegisctl configuration file structure.
type Config struct {
	CurrentContext string             `yaml:"current-context"`
	Contexts       map[string]Context `yaml:"contexts"`
	Preferences    Preferences        `yaml:"preferences,omitempty"`
}

// Context represents a named connection context.
type Context struct {
	Server         string    `yaml:"server"`
	Token          string    `yaml:"token,omitempty"`
	AuthType       string    `yaml:"auth-type,omitempty"`
	AccessKey      string    `yaml:"access-key,omitempty"`
	DefaultProject string    `yaml:"default-project,omitempty"`
	TokenExpiry    time.Time `yaml:"token-expiry,omitempty"`
}

// Preferences holds user-level defaults.
type Preferences struct {
	Output         string `yaml:"output,omitempty"`
	RequestTimeout int    `yaml:"request-timeout,omitempty"`
}

// configDir returns ~/.aegisctl, creating it if necessary.
func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	dir := filepath.Join(home, ".aegisctl")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("cannot create config directory: %w", err)
	}
	return dir, nil
}

// configPath returns the full path to the config file.
func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// LoadConfig reads the config file from disk. If the file does not exist a
// default empty config is returned without error.
func LoadConfig() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Contexts: make(map[string]Context),
			}, nil
		}
		return nil, fmt.Errorf("cannot read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("cannot parse config: %w", err)
	}
	if cfg.Contexts == nil {
		cfg.Contexts = make(map[string]Context)
	}
	return &cfg, nil
}

// SaveConfig writes the config to disk.
func SaveConfig(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("cannot write config: %w", err)
	}
	return nil
}

// GetCurrentContext returns the active context from the config. If no current
// context is set an error is returned.
func GetCurrentContext(cfg *Config) (*Context, string, error) {
	if cfg.CurrentContext == "" {
		return nil, "", fmt.Errorf("no current context set; run 'aegisctl context use <name>' or 'aegisctl auth login'")
	}
	ctx, ok := cfg.Contexts[cfg.CurrentContext]
	if !ok {
		return nil, "", fmt.Errorf("current context %q not found in config", cfg.CurrentContext)
	}
	return &ctx, cfg.CurrentContext, nil
}

// SetCurrentContext switches the active context.
func SetCurrentContext(cfg *Config, name string) error {
	if _, ok := cfg.Contexts[name]; !ok {
		return fmt.Errorf("context %q does not exist", name)
	}
	cfg.CurrentContext = name
	return nil
}
