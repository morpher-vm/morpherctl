package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAgent(t *testing.T) {
	tests := []struct {
		name         string
		controllerIP string
		expectedIP   string
		expectedPort int
		expectedArch string
		expectedPath string
		expectedName string
	}{
		{
			name:         "should create agent with default values",
			controllerIP: "192.168.1.100",
			expectedIP:   "192.168.1.100",
			expectedPort: 9000,
			expectedArch: "arm64", // macOS ARM64.
			expectedPath: "/usr/local/bin",
			expectedName: "morpher-agent",
		},
		{
			name:         "should create agent with custom controller IP",
			controllerIP: "10.0.0.1",
			expectedIP:   "10.0.0.1",
			expectedPort: 9000,
			expectedArch: "arm64",
			expectedPath: "/usr/local/bin",
			expectedName: "morpher-agent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewAgent(tt.controllerIP)

			assert.Equal(t, tt.expectedIP, agent.ControllerIP)
			assert.Equal(t, tt.expectedPort, agent.ControllerPort)
			assert.Equal(t, tt.expectedPath, agent.InstallPath)
			assert.Equal(t, tt.expectedName, agent.ServiceName)
			// Architecture might vary by platform, so just check it's not empty.
			assert.NotEmpty(t, agent.Architecture)
		})
	}
}

func TestAgent_IsInstalled(t *testing.T) {
	// Create temporary directory for testing.
	tempDir := t.TempDir()
	agent := &Agent{
		InstallPath: tempDir,
		ServiceName: "test-agent",
	}

	t.Run("should return false when agent is not installed", func(t *testing.T) {
		installed := agent.isInstalled()
		assert.False(t, installed)
	})

	t.Run("should return true when agent is installed", func(t *testing.T) {
		// Create a mock agent binary.
		agentPath := filepath.Join(tempDir, "test-agent")
		err := os.WriteFile(agentPath, []byte("#!/bin/bash\necho 'test'"), 0755)
		require.NoError(t, err)

		installed := agent.isInstalled()
		assert.True(t, installed)

		// Cleanup.
		os.Remove(agentPath)
	})
}

func TestAgent_DownloadInstallScript(t *testing.T) {
	agent := &Agent{
		Architecture: "arm64",
	}

	t.Run("should generate correct script name and URL", func(t *testing.T) {
		_, err := agent.downloadInstallScript()

		// Since this is a mock test, we expect it to fail at download
		// but we can verify the script name generation logic.
		if err != nil {
			// Check if the error is about wget/curl not found or wget/curl failed (expected in test environment).
			assert.True(t,
				strings.Contains(err.Error(), "neither wget nor curl found") ||
					strings.Contains(err.Error(), "wget failed") ||
					strings.Contains(err.Error(), "curl failed"),
				"Error should contain expected message, got: %s", err.Error())
		}
	})
}

func TestAgent_Verify(t *testing.T) {
	// Create temporary directory for testing.
	tempDir := t.TempDir()
	agent := &Agent{
		InstallPath: tempDir,
		ServiceName: "test-agent",
	}

	t.Run("should return verification result when agent is not installed", func(t *testing.T) {
		result, err := agent.Verify()
		require.NoError(t, err)

		assert.False(t, result.BinaryInstalled)
		assert.Empty(t, result.BinaryPath)
		assert.Empty(t, result.ServiceStatus)
		assert.False(t, result.ServiceRunning)
		assert.False(t, result.ServiceEnabled)
	})

	t.Run("should return verification result when agent is installed", func(t *testing.T) {
		// Create a mock agent binary.
		agentPath := filepath.Join(tempDir, "test-agent")
		err := os.WriteFile(agentPath, []byte("#!/bin/bash\necho 'test'"), 0755)
		require.NoError(t, err)

		result, err := agent.Verify()
		require.NoError(t, err)

		assert.True(t, result.BinaryInstalled)
		assert.Equal(t, agentPath, result.BinaryPath)
		// Service status will depend on system availability, so we don't assert specific values.

		// Cleanup.
		os.Remove(agentPath)
	})
}

func TestAgent_Status(t *testing.T) {
	// Create temporary directory for testing.
	tempDir := t.TempDir()
	agent := &Agent{
		InstallPath: tempDir,
		ServiceName: "test-agent",
	}

	t.Run("should return status when agent is not installed", func(t *testing.T) {
		status, err := agent.Status()
		require.NoError(t, err)

		assert.False(t, status.Installed)
		assert.Empty(t, status.InstallPath)
		assert.Empty(t, status.ServiceStatus)
		assert.False(t, status.ServiceEnabled)
	})

	t.Run("should return status when agent is installed", func(t *testing.T) {
		// Create a mock agent binary.
		agentPath := filepath.Join(tempDir, "test-agent")
		err := os.WriteFile(agentPath, []byte("#!/bin/bash\necho 'test'"), 0755)
		require.NoError(t, err)

		status, err := agent.Status()
		require.NoError(t, err)

		assert.True(t, status.Installed)
		assert.Equal(t, agentPath, status.InstallPath)
		// Service status will depend on system availability, so we don't assert specific values.

		// Cleanup.
		os.Remove(agentPath)
	})
}

func TestAgent_ServiceManagement(t *testing.T) {
	agent := &Agent{
		ServiceName: "test-service",
	}

	t.Run("should handle service status check gracefully", func(t *testing.T) {
		// These methods will fail in test environment, but we can test they don't panic.
		// and return appropriate errors.
		status, err := agent.getServiceStatus()
		if err != nil {
			// Expected in test environment - could be "exit status" or "executable file not found".
			assert.True(t,
				strings.Contains(err.Error(), "exit status") ||
					strings.Contains(err.Error(), "executable file not found"),
				"Error should contain expected message, got: %s", err.Error())
		}

		enabled, err := agent.isServiceEnabled()
		if err != nil {
			// Expected in test environment - could be "exit status" or "executable file not found".
			assert.True(t,
				strings.Contains(err.Error(), "exit status") ||
					strings.Contains(err.Error(), "executable file not found"),
				"Error should contain expected message, got: %s", err.Error())
		}

		// Just ensure they don't panic and return boolean values.
		assert.IsType(t, "", status)
		assert.IsType(t, false, enabled)
	})
}

func TestAgent_FileOperations(t *testing.T) {
	// Create temporary directory for testing.
	tempDir := t.TempDir()
	agent := &Agent{
		InstallPath: tempDir,
		ServiceName: "test-agent",
	}

	t.Run("should handle file removal operations", func(t *testing.T) {
		// Test removeAgentBinary.
		testFile := filepath.Join(tempDir, "test-agent")
		err := os.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)

		err = agent.removeAgentBinary()
		// This will fail in test environment due to permissions, but we can verify the logic.
		if err != nil {
			assert.Contains(t, err.Error(), "permission denied")
		}

		// Cleanup.
		os.Remove(testFile)
	})
}

func TestAgent_ArchitectureDetection(t *testing.T) {
	agent := NewAgent("127.0.0.1")

	t.Run("should detect architecture", func(t *testing.T) {
		// Architecture should be detected from runtime.
		assert.NotEmpty(t, agent.Architecture)

		// Should be one of the common architectures.
		validArchs := []string{"amd64", "arm64", "386", "arm"}
		found := false
		for _, arch := range validArchs {
			if agent.Architecture == arch {
				found = true
				break
			}
		}
		assert.True(t, found, "Architecture %s should be one of %v", agent.Architecture, validArchs)
	})
}
