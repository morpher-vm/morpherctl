package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name        string
		configFile  string
		expectedDir string
	}{
		{
			name:        "empty config file should use default path",
			configFile:  "",
			expectedDir: ".morpherctl",
		},
		{
			name:        "custom config file should be preserved",
			configFile:  "/custom/path/config.yaml",
			expectedDir: "/custom/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager(tt.configFile)

			if tt.configFile == "" {
				// Check that default path contains .morpherctl.
				assert.Contains(t, manager.GetConfigFile(), ".morpherctl")
				assert.Contains(t, manager.GetConfigDir(), ".morpherctl")
			} else {
				assert.Equal(t, tt.configFile, manager.GetConfigFile())
				// For custom config files, configDir should be the directory of the config file.
				expectedDir := filepath.Dir(tt.configFile)
				assert.Equal(t, expectedDir, manager.GetConfigDir())
			}
		})
	}
}

func TestManager_Init(t *testing.T) {
	// Create temporary directory for testing.
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.yaml")

	manager := NewManager(configFile)

	t.Run("should initialize configuration successfully", func(t *testing.T) {
		err := manager.Init()
		require.NoError(t, err)

		// Check if config file exists.
		_, err = os.Stat(configFile)
		require.NoError(t, err)

		// Check if config directory exists.
		_, err = os.Stat(tempDir)
		require.NoError(t, err)
	})

	t.Run("should set default values", func(t *testing.T) {
		// Reload the manager to read the created config.
		manager = NewManager(configFile)

		// Test default values.
		url, err := manager.GetString("controller.url")
		require.NoError(t, err)
		assert.Equal(t, "http://localhost:8080", url)

		timeout, err := manager.GetDuration("controller.timeout")
		require.NoError(t, err)
		assert.Equal(t, 30*time.Second, timeout)

		installPath, err := manager.GetString("agent.install_path")
		require.NoError(t, err)
		assert.Equal(t, "/opt/morpher", installPath)

		logLevel, err := manager.GetString("agent.log_level")
		require.NoError(t, err)
		assert.Equal(t, "info", logLevel)
	})
}

func TestManager_SetAndGet(t *testing.T) {
	// Create temporary directory for testing.
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.yaml")

	manager := NewManager(configFile)

	// Initialize first.
	err := manager.Init()
	require.NoError(t, err)

	t.Run("should set and get string value", func(t *testing.T) {
		err := manager.Set("test.key", "test_value")
		require.NoError(t, err)

		value, err := manager.Get("test.key")
		require.NoError(t, err)
		assert.Equal(t, "test_value", value)
	})

	t.Run("should set and get multiple values", func(t *testing.T) {
		testData := map[string]string{
			"test.key1": "value1",
			"test.key2": "value2",
			"test.key3": "value3",
		}

		for key, value := range testData {
			err := manager.Set(key, value)
			require.NoError(t, err)
		}

		for key, expectedValue := range testData {
			value, err := manager.Get(key)
			require.NoError(t, err)
			assert.Equal(t, expectedValue, value)
		}
	})

	t.Run("should return error for non-existent key", func(t *testing.T) {
		_, err := manager.Get("non.existent.key")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestManager_GetAll(t *testing.T) {
	// Create temporary directory for testing.
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.yaml")

	manager := NewManager(configFile)

	// Initialize first.
	err := manager.Init()
	require.NoError(t, err)

	// Set some test values.
	err = manager.Set("test.key1", "value1")
	require.NoError(t, err)
	err = manager.Set("test.key2", "value2")
	require.NoError(t, err)

	t.Run("should return all configuration values", func(t *testing.T) {
		allConfig, err := manager.GetAll()
		require.NoError(t, err)

		// Debug: print the actual structure.
		t.Logf("All config: %+v", allConfig)

		// Should contain our test values in nested structure.
		assert.Contains(t, allConfig, "test")

		// Check the nested test values.
		testSection, exists := allConfig["test"]
		assert.True(t, exists)

		testMap, ok := testSection.(map[string]any)
		assert.True(t, ok)

		assert.Contains(t, testMap, "key1")
		assert.Contains(t, testMap, "key2")
		assert.Equal(t, "value1", testMap["key1"])
		assert.Equal(t, "value2", testMap["key2"])
	})
}

func TestManager_GetString(t *testing.T) {
	// Create temporary directory for testing.
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.yaml")

	manager := NewManager(configFile)

	// Initialize first.
	err := manager.Init()
	require.NoError(t, err)

	t.Run("should return string value", func(t *testing.T) {
		url, err := manager.GetString("controller.url")
		require.NoError(t, err)
		assert.Equal(t, "http://localhost:8080", url)
	})

	t.Run("should return empty string for non-existent key", func(t *testing.T) {
		value, err := manager.GetString("non.existent.key")
		require.NoError(t, err)
		assert.Equal(t, "", value)
	})
}

func TestManager_GetDuration(t *testing.T) {
	// Create temporary directory for testing.
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.yaml")

	manager := NewManager(configFile)

	// Initialize first.
	err := manager.Init()
	require.NoError(t, err)

	t.Run("should return duration value", func(t *testing.T) {
		timeout, err := manager.GetDuration("controller.timeout")
		require.NoError(t, err)
		assert.Equal(t, 30*time.Second, timeout)
	})
}

func TestManager_LoadError(t *testing.T) {
	// Create manager with non-existent config file.
	manager := NewManager("/non/existent/config.yaml")

	t.Run("should return error when loading non-existent config", func(t *testing.T) {
		_, err := manager.Get("any.key")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read configuration file")
	})
}
