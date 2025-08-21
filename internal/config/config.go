package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// Manager handles configuration operations.
type Manager struct {
	configFile string
	configDir  string
}

// NewManager creates a new configuration manager.
func NewManager(configFile string) *Manager {
	// Set default configuration directory.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	defaultConfigDir := filepath.Join(homeDir, ".morpherctl")

	// Set configuration file path if not provided.
	if configFile == "" {
		configFile = filepath.Join(defaultConfigDir, "config.yaml")
	}

	// Determine the actual config directory based on the config file path.
	// If a custom config file path is provided, use its directory.
	// Otherwise, use the default directory.
	configDir := filepath.Dir(configFile)
	if configFile == filepath.Join(defaultConfigDir, "config.yaml") {
		configDir = defaultConfigDir
	}

	return &Manager{
		configFile: configFile,
		configDir:  configDir,
	}
}

// Init initializes the configuration file with default values.
func (m *Manager) Init() error {
	// Create configuration directory for the config file.
	// This ensures the directory exists even for custom config file paths.
	configFileDir := filepath.Dir(m.configFile)
	if err := os.MkdirAll(configFileDir, 0755); err != nil {
		return fmt.Errorf("failed to create configuration directory: %w", err)
	}

	// Set default configuration values.
	viper.SetDefault("controller.ip", "localhost")
	viper.SetDefault("controller.port", 9000)
	viper.SetDefault("controller.timeout", "30s")
	viper.SetDefault("auth.token", "")
	viper.SetDefault("auth.refresh_token", "")
	viper.SetDefault("agent.install_path", "/opt/morpher")
	viper.SetDefault("agent.log_level", "info")

	// Set configuration file path.
	viper.SetConfigFile(m.configFile)
	viper.SetConfigType("yaml")

	// Save configuration file.
	if err := viper.WriteConfigAs(m.configFile); err != nil {
		return fmt.Errorf("failed to save configuration file: %w", err)
	}

	return nil
}

// Set sets a configuration key-value pair.
func (m *Manager) Set(key, value string) error {
	// Load configuration file.
	if err := m.load(); err != nil {
		return err
	}

	// Set configuration value.
	viper.Set(key, value)

	// Save configuration file.
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

// Get retrieves a configuration value by key.
func (m *Manager) Get(key string) (any, error) {
	// Load configuration file.
	if err := m.load(); err != nil {
		return nil, err
	}

	value := viper.Get(key)
	if value == nil {
		return nil, fmt.Errorf("configuration key '%s' not found", key)
	}

	return value, nil
}

// GetAll retrieves all configuration values.
func (m *Manager) GetAll() (map[string]any, error) {
	// Load configuration file.
	if err := m.load(); err != nil {
		return nil, err
	}

	return viper.AllSettings(), nil
}

// GetString retrieves a string configuration value by key.
func (m *Manager) GetString(key string) (string, error) {
	// Load configuration file.
	if err := m.load(); err != nil {
		return "", err
	}

	return viper.GetString(key), nil
}

// GetDuration retrieves a duration configuration value by key.
func (m *Manager) GetDuration(key string) (time.Duration, error) {
	// Load configuration file.
	if err := m.load(); err != nil {
		return 0, err
	}

	return viper.GetDuration(key), nil
}

// load loads the configuration file.
func (m *Manager) load() error {
	viper.SetConfigFile(m.configFile)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read configuration file: %w", err)
	}

	return nil
}

// GetConfigFile returns the current configuration file path.
func (m *Manager) GetConfigFile() string {
	return m.configFile
}

// GetConfigDir returns the current configuration directory path.
func (m *Manager) GetConfigDir() string {
	return m.configDir
}
