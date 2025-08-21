package agent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// VerificationResult represents the result of agent verification.
type VerificationResult struct {
	BinaryInstalled bool
	BinaryPath      string
	ServiceStatus   string
	ServiceRunning  bool
	ServiceEnabled  bool
}

// StatusResult represents the current status of the agent.
type StatusResult struct {
	Installed      bool
	InstallPath    string
	ServiceStatus  string
	ServiceEnabled bool
}

// Agent represents a morpher agent installation.
type Agent struct {
	ControllerIP   string
	ControllerPort int
	Architecture   string
	InstallPath    string
	ServiceName    string
}

// NewAgent creates a new Agent instance.
func NewAgent(controllerIP string) *Agent {
	return &Agent{
		ControllerIP:   controllerIP,
		ControllerPort: 9000,
		Architecture:   runtime.GOARCH,
		InstallPath:    "/usr/local/bin",
		ServiceName:    "morpher-agent",
	}
}

// Install downloads and installs the morpher agent.
func (a *Agent) Install(_ string) error {
	// Download installation script.
	scriptPath, err := a.downloadInstallScript()
	if err != nil {
		return fmt.Errorf("failed to download install script: %w", err)
	}

	// Make script executable.
	if err := os.Chmod(scriptPath, 0755); err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	// Run installation script.
	if err := a.runInstallScript(scriptPath); err != nil {
		return fmt.Errorf("failed to run install script: %w", err)
	}

	// Clean up script.
	os.Remove(scriptPath)

	return nil
}

// Uninstall removes the morpher agent.
func (a *Agent) Uninstall() error {
	// Stop and disable service.
	if err := a.stopService(); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	// Remove service file.
	if err := a.removeServiceFile(); err != nil {
		return fmt.Errorf("failed to remove service file: %w", err)
	}

	// Remove agent binary.
	if err := a.removeAgentBinary(); err != nil {
		return fmt.Errorf("failed to remove agent binary: %w", err)
	}

	// Remove config directory.
	if err := a.removeConfigDirectory(); err != nil {
		return fmt.Errorf("failed to remove config directory: %w", err)
	}

	// Reload systemd.
	if err := a.reloadSystemd(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	return nil
}

// Upgrade upgrades the morpher agent to a new version.
func (a *Agent) Upgrade(version string) error {
	// Check if agent is installed.
	if !a.isInstalled() {
		return fmt.Errorf("agent is not installed")
	}

	// Stop service.
	if err := a.stopService(); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	// Install new version.
	if err := a.Install(version); err != nil {
		return fmt.Errorf("failed to install new version: %w", err)
	}

	// Start service.
	if err := a.startService(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	return nil
}

// Verify checks the agent installation and service status.
func (a *Agent) Verify() (*VerificationResult, error) {
	result := &VerificationResult{}

	// Check if binary exists.
	if a.isInstalled() {
		result.BinaryInstalled = true
		result.BinaryPath = filepath.Join(a.InstallPath, a.ServiceName)
	}

	// Check service status.
	if status, err := a.getServiceStatus(); err == nil {
		result.ServiceStatus = status
		result.ServiceRunning = status == "active"
	}

	// Check if service is enabled.
	if enabled, err := a.isServiceEnabled(); err == nil {
		result.ServiceEnabled = enabled
	}

	return result, nil
}

// Status returns the current status of the agent.
func (a *Agent) Status() (*StatusResult, error) {
	status := &StatusResult{}

	// Check installation.
	if a.isInstalled() {
		status.Installed = true
		status.InstallPath = filepath.Join(a.InstallPath, a.ServiceName)
	}

	// Check service status.
	if serviceStatus, err := a.getServiceStatus(); err == nil {
		status.ServiceStatus = serviceStatus
	}

	// Check if service is enabled.
	if enabled, err := a.isServiceEnabled(); err == nil {
		status.ServiceEnabled = enabled
	}

	return status, nil
}

// Helper methods.
func (a *Agent) downloadInstallScript() (string, error) {
	scriptName := fmt.Sprintf("install_%s.sh", a.Architecture)
	scriptURL := fmt.Sprintf("https://raw.githubusercontent.com/morpher-vm/morpher-agent/main/scripts/%s", scriptName)

	// Create temporary file.
	tmpFile, err := os.CreateTemp("", scriptName)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	// Download script using wget or curl.
	if err := a.downloadFile(scriptURL, tmpFile.Name()); err != nil {
		return "", fmt.Errorf("failed to download script: %w", err)
	}

	return tmpFile.Name(), nil
}

func (a *Agent) downloadFile(url, dest string) error {
	// Try wget first.
	if _, err := exec.LookPath("wget"); err == nil {
		cmd := exec.Command("wget", "-O", dest, url)
		return fmt.Errorf("wget failed: %w", cmd.Run())
	}

	// Fallback to curl.
	if _, err := exec.LookPath("curl"); err == nil {
		cmd := exec.Command("curl", "-o", dest, url)
		return fmt.Errorf("curl failed: %w", cmd.Run())
	}

	return fmt.Errorf("neither wget nor curl found")
}

func (a *Agent) runInstallScript(scriptPath string) error {
	cmd := exec.Command("sudo", "-E", scriptPath)
	cmd.Env = append(os.Environ(), fmt.Sprintf("MORPHER_CONTROLLER_IP=%s", a.ControllerIP))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run install script: %w", err)
	}
	return nil
}

func (a *Agent) isInstalled() bool {
	_, err := os.Stat(filepath.Join(a.InstallPath, a.ServiceName))
	return err == nil
}

func (a *Agent) stopService() error {
	cmd := exec.Command("sudo", "systemctl", "disable", "--now", a.ServiceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}
	return nil
}

func (a *Agent) startService() error {
	cmd := exec.Command("sudo", "systemctl", "start", a.ServiceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}
	return nil
}

func (a *Agent) removeServiceFile() error {
	servicePath := fmt.Sprintf("/etc/systemd/system/%s.service", a.ServiceName)
	if err := os.Remove(servicePath); err != nil {
		return fmt.Errorf("failed to remove service file: %w", err)
	}
	return nil
}

func (a *Agent) removeAgentBinary() error {
	binaryPath := filepath.Join(a.InstallPath, a.ServiceName)
	if err := os.Remove(binaryPath); err != nil {
		return fmt.Errorf("failed to remove agent binary: %w", err)
	}
	return nil
}

func (a *Agent) removeConfigDirectory() error {
	configPath := "/etc/morpher-agent"
	if err := os.RemoveAll(configPath); err != nil {
		return fmt.Errorf("failed to remove config directory: %w", err)
	}
	return nil
}

func (a *Agent) reloadSystemd() error {
	cmd := exec.Command("sudo", "systemctl", "daemon-reload")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}
	return nil
}

func (a *Agent) getServiceStatus() (string, error) {
	cmd := exec.Command("systemctl", "is-active", a.ServiceName)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get service status: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func (a *Agent) isServiceEnabled() (bool, error) {
	cmd := exec.Command("systemctl", "is-enabled", a.ServiceName)
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check if service is enabled: %w", err)
	}
	return strings.TrimSpace(string(output)) == "enabled", nil
}
