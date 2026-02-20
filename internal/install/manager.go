package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"atlas.hub/internal/model"
	"github.com/go-git/go-git/v5"
)

type Manager struct {
	InstallPath string
	TempDir     string
	GobakePath  string
}

func NewManager(installPath string) (*Manager, error) {
	tempDir, err := os.MkdirTemp("", "atlas-hub-*")
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(installPath, 0755); err != nil {
		return nil, err
	}

	m := &Manager{
		InstallPath: installPath,
		TempDir:     tempDir,
	}

	path, err := m.ensureGobake()
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("gobake requirement failed: %w", err)
	}
	m.GobakePath = path

	return m, nil
}

func (m *Manager) Cleanup() {
	os.RemoveAll(m.TempDir)
}

func (m *Manager) ensureGobake() (string, error) {
	path, err := exec.LookPath("gobake")
	if err == nil {
		return path, nil
	}

	// Try go install
	cmd := exec.Command("go", "install", "github.com/fezcode/gobake@latest")
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("install failed: %s", string(out))
	}

	// Check GOPATH/bin
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		home, _ := os.UserHomeDir()
		gopath = filepath.Join(home, "go")
	}
	binPath := filepath.Join(gopath, "bin", "gobake")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}

	if _, err := os.Stat(binPath); err == nil {
		return binPath, nil
	}

	return "", fmt.Errorf("gobake installed but not found in PATH or GOPATH/bin")
}

func (m *Manager) Install(tool *model.Tool) error {
	toolDir := filepath.Join(m.TempDir, tool.Name)
	
	// 1. Clone
	_, err := git.PlainClone(toolDir, false, &git.CloneOptions{
		URL:      tool.Repo,
		Progress: nil,
	})
	if err != nil {
		return fmt.Errorf("clone failed: %w", err)
	}

	// 2. Build with gobake
	cmd := exec.Command(m.GobakePath, "build")
	cmd.Dir = toolDir
	// Set env to ensure CGO is correct if needed, but inheriting env is usually fine
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("build failed: %s", string(output))
	}

	// 3. Find binary
	buildDir := filepath.Join(toolDir, "build")
	entries, err := os.ReadDir(buildDir)
	if err != nil {
		return fmt.Errorf("read build dir failed: %w", err)
	}

	var binaryPath string
	target := fmt.Sprintf("%s-%s-%s", tool.Name, runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		target += ".exe"
	}

	for _, e := range entries {
		if e.Name() == target {
			binaryPath = filepath.Join(buildDir, e.Name())
			break
		}
	}

	if binaryPath == "" {
		return fmt.Errorf("binary not found for target %s", target)
	}

	// 4. Move to install path
	destName := tool.Bin
	if runtime.GOOS == "windows" && !strings.HasSuffix(destName, ".exe") {
		destName += ".exe"
	}
	destPath := filepath.Join(m.InstallPath, destName)

	// Remove existing if any
	os.Remove(destPath)

	// Copy/Move
	input, err := os.ReadFile(binaryPath)
	if err != nil {
		return err
	}
	if err := os.WriteFile(destPath, input, 0755); err != nil {
		return err
	}

	return nil
}
