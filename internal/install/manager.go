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
	if err != nil {
		// Try go install if not found
		cmd := exec.Command("go", "install", "github.com/fezcode/gobake@latest")
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("gobake install failed: %s", string(out))
		}
		
		// Re-check path
		path, err = exec.LookPath("gobake")
		if err != nil {
			// Try GOPATH/bin as last resort
			gopath := os.Getenv("GOPATH")
			if gopath == "" {
				home, _ := os.UserHomeDir()
				gopath = filepath.Join(home, "go")
			}
			binPath := filepath.Join(gopath, "bin", "gobake")
			if runtime.GOOS == "windows" {
				binPath += ".exe"
			}
			if _, statErr := os.Stat(binPath); statErr == nil {
				path = binPath
			} else {
				return "", fmt.Errorf("gobake not found in PATH or GOPATH/bin after install")
			}
		}
	}
	return path, nil
}

func (m *Manager) Install(tool *model.Tool, onProgress func(string)) error {
	toolDir := filepath.Join(m.TempDir, tool.Name)
	
	if onProgress != nil { onProgress("Cloning repository...") }
	// 1. Clone
	_, err := git.PlainClone(toolDir, false, &git.CloneOptions{
		URL:      tool.Repo,
		Progress: nil,
	})
	if err != nil {
		return fmt.Errorf("clone failed: %w", err)
	}

	if onProgress != nil { onProgress("Resolving dependencies...") }
	// 2. Tidy dependencies (Fixes compass error)
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = toolDir
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy failed: %s", string(output))
	}

	if onProgress != nil { onProgress("Building binary...") }
	// 3. Build with gobake
	buildCmd := exec.Command(m.GobakePath, "build")
	buildCmd.Dir = toolDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("build failed: %s", string(output))
	}

	// 4. Find binary (Flexible detection for games/wilson-revenge etc)
	buildDir := filepath.Join(toolDir, "build")
	entries, err := os.ReadDir(buildDir)
	if err != nil {
		return fmt.Errorf("read build dir failed: %w", err)
	}

	var binaryPath string
	suffix := fmt.Sprintf("-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		suffix += ".exe"
	}

	for _, e := range entries {
		if strings.HasSuffix(e.Name(), suffix) {
			binaryPath = filepath.Join(buildDir, e.Name())
			break
		}
	}

	if binaryPath == "" {
		return fmt.Errorf("binary with suffix %s not found in build dir", suffix)
	}

	if onProgress != nil { onProgress("Installing...") }
	// 5. Move to install path
	destName := tool.Bin
	if runtime.GOOS == "windows" && !strings.HasSuffix(destName, ".exe") {
		destName += ".exe"
	}
	destPath := filepath.Join(m.InstallPath, destName)

	// Safe update (especially for Windows): 
	// If the file exists, rename it to .old before writing the new one.
	// Windows allows renaming running binaries, but not overwriting.
	if _, err := os.Stat(destPath); err == nil {
		oldPath := destPath + ".old"
		os.Remove(oldPath) // Remove previous .old if any
		if err := os.Rename(destPath, oldPath); err != nil {
			// If rename fails (might be okay if it's not the running process), 
			// try to remove it as a fallback.
			os.Remove(destPath)
		}
	}

	// Copy/Move binary to destination
	input, err := os.ReadFile(binaryPath)
	if err != nil {
		return err
	}
	if err := os.WriteFile(destPath, input, 0755); err != nil {
		return err
	}

	return nil
}
