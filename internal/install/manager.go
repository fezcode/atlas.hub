package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"atlas.hub/internal/model"
	"github.com/go-git/go-git/v5"
)

type Manager struct {
	InstallPath string
	TempDir     string
	GobakePath  string
	LogDir      string
}

func NewManager(installPath string) (*Manager, error) {
	tempDir, err := os.MkdirTemp("", "atlas-hub-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	if err := os.MkdirAll(installPath, 0755); err != nil {
		return nil, fmt.Errorf("create install dir %q: %w", installPath, err)
	}

	logDir := ""
	if home, herr := os.UserHomeDir(); herr == nil {
		logDir = filepath.Join(home, ".atlas", "hub-data", "logs")
		_ = os.MkdirAll(logDir, 0755)
	}

	m := &Manager{
		InstallPath: installPath,
		TempDir:     tempDir,
		LogDir:      logDir,
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
	if _, err := exec.LookPath("go"); err != nil {
		return "", fmt.Errorf("`go` toolchain not found in PATH: %w\n    → install Go from https://go.dev/dl/ and ensure `go` is on PATH, then retry", err)
	}

	path, err := exec.LookPath("gobake")
	if err != nil {
		installSpec := "github.com/fezcode/gobake/cmd/gobake@latest"
		cmd := exec.Command("go", "install", installSpec)
		out, runErr := cmd.CombinedOutput()
		if runErr != nil {
			return "", fmt.Errorf("bootstrap `go install %s` failed (exit: %v)\n    stdout+stderr:\n%s",
				installSpec, runErr, indent(string(out)))
		}

		path, err = exec.LookPath("gobake")
		if err != nil {
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
				return "", fmt.Errorf("gobake binary not found after `go install` — looked in PATH and %q\n    → add %s to PATH, or check `go env GOBIN` / `go env GOPATH`",
					binPath, filepath.Dir(binPath))
			}
		}
	}
	return path, nil
}

func indent(s string) string {
	if s == "" {
		return "      (no output)"
	}
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i, l := range lines {
		lines[i] = "      " + l
	}
	return strings.Join(lines, "\n")
}

func (m *Manager) Install(tool *model.Tool) error {
	toolDir := filepath.Join(m.TempDir, tool.Name)
	logPath := m.startLog(tool.Name)
	m.logf(logPath, "install start: tool=%s repo=%s workdir=%s go=%s arch=%s",
		tool.Name, tool.Repo, toolDir, runtime.GOOS, runtime.GOARCH)

	// 1. Clone
	m.logf(logPath, "[clone] git clone %s -> %s", tool.Repo, toolDir)
	if _, err := git.PlainClone(toolDir, false, &git.CloneOptions{
		URL:      tool.Repo,
		Progress: nil,
	}); err != nil {
		m.logf(logPath, "[clone] FAILED: %v", err)
		return fmt.Errorf("git clone %s failed: %w\n    → check network / proxy, and that the repo is public\n    log: %s",
			tool.Repo, err, logPath)
	}

	// 2. Tidy dependencies
	m.logf(logPath, "[tidy] go mod tidy (cwd=%s)", toolDir)
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = toolDir
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		m.logf(logPath, "[tidy] FAILED: %v\n%s", err, string(output))
		return fmt.Errorf("`go mod tidy` failed in %s (exit: %v)\n    stdout+stderr:\n%s\n    → try `atlas.hub --clear-go-cache` then retry; check internet / GOPROXY\n    log: %s",
			toolDir, err, indent(string(output)), logPath)
	}

	// 3. Build with gobake
	m.logf(logPath, "[build] %s build (cwd=%s)", m.GobakePath, toolDir)
	buildCmd := exec.Command(m.GobakePath, "build")
	buildCmd.Dir = toolDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		m.logf(logPath, "[build] FAILED: %v\n%s", err, string(output))
		return fmt.Errorf("`%s build` failed in %s (exit: %v)\n    stdout+stderr:\n%s\n    → if this is a CGO project (e.g. atlas.pilot), install a C compiler (MinGW-w64/gcc) and required native libs\n    → try `atlas.hub --clear-go-cache` if the cache looks stale\n    log: %s",
			filepath.Base(m.GobakePath), toolDir, err, indent(string(output)), logPath)
	}

	// 4. Find binary
	buildDir := filepath.Join(toolDir, "build")
	entries, err := os.ReadDir(buildDir)
	if err != nil {
		m.logf(logPath, "[find] read %s FAILED: %v", buildDir, err)
		return fmt.Errorf("reading build output dir %s failed: %w\n    → the recipe's `build` task may have skipped producing any artifact\n    log: %s",
			buildDir, err, logPath)
	}

	var binaryPath string
	suffix := fmt.Sprintf("-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		suffix += ".exe"
	}

	var present []string
	for _, e := range entries {
		present = append(present, e.Name())
		if strings.HasSuffix(e.Name(), suffix) {
			binaryPath = filepath.Join(buildDir, e.Name())
			break
		}
	}

	if binaryPath == "" {
		listed := "(empty)"
		if len(present) > 0 {
			listed = strings.Join(present, ", ")
		}
		m.logf(logPath, "[find] no %s match in %s; present: %s", suffix, buildDir, listed)
		return fmt.Errorf("no binary matching suffix %q found in %s\n    files present: %s\n    → the `gobake build` command exited 0 but produced no artifact for your OS/arch; check the tool's Recipe.go `targets` list\n    → for CGO tools, the build may have silently skipped due to a missing compiler — re-run `gobake build` manually in %s to see detail\n    log: %s",
			suffix, buildDir, listed, toolDir, logPath)
	}

	// 5. Move to install path
	destName := tool.Bin
	if runtime.GOOS == "windows" && !strings.HasSuffix(destName, ".exe") {
		destName += ".exe"
	}
	destPath := filepath.Join(m.InstallPath, destName)

	// Safe update (timestamped backup)
	if _, err := os.Stat(destPath); err == nil {
		oldPath := fmt.Sprintf("%s.%d.old", destPath, time.Now().UnixNano())
		if err := os.Rename(destPath, oldPath); err != nil {
			os.Remove(destPath)
		} else {
			matches, _ := filepath.Glob(destPath + ".*.old")
			for _, match := range matches {
				os.Remove(match)
			}
		}
	}

	// Copy/Move binary
	input, err := os.ReadFile(binaryPath)
	if err != nil {
		m.logf(logPath, "[install] read %s FAILED: %v", binaryPath, err)
		return fmt.Errorf("reading built binary %s failed: %w\n    log: %s", binaryPath, err, logPath)
	}
	if err := os.WriteFile(destPath, input, 0755); err != nil {
		m.logf(logPath, "[install] write %s FAILED: %v", destPath, err)
		return fmt.Errorf("writing binary to %s failed: %w\n    → is the file currently running, or is the directory read-only / on a different volume?\n    log: %s",
			destPath, err, logPath)
	}
	m.logf(logPath, "[install] ok -> %s", destPath)

	return nil
}

func (m *Manager) startLog(toolName string) string {
	if m.LogDir == "" {
		return ""
	}
	name := fmt.Sprintf("%s-%s.log", toolName, time.Now().Format("20060102-150405"))
	return filepath.Join(m.LogDir, name)
}

func (m *Manager) logf(path, format string, args ...any) {
	if path == "" {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "[%s] %s\n", time.Now().Format("15:04:05.000"), fmt.Sprintf(format, args...))
}

func (m *Manager) Delete(tool *model.Tool) error {
	destName := tool.Bin
	if runtime.GOOS == "windows" && !strings.HasSuffix(destName, ".exe") {
		destName += ".exe"
	}
	return os.Remove(filepath.Join(m.InstallPath, destName))
}

func (m *Manager) CheckInstalledVersion(tool *model.Tool) {
	destName := tool.Bin
	if runtime.GOOS == "windows" && !strings.HasSuffix(destName, ".exe") {
		destName += ".exe"
	}
	binPath := filepath.Join(m.InstallPath, destName)

	if _, err := os.Stat(binPath); err != nil {
		return // not installed
	}

	// Try to read the version from the binary; fall back to the manifest version
	// so tools that don't output a clean --version still show as installed.
	cmd := exec.Command(binPath, "--version")
	out, err := cmd.Output()
	if err == nil {
		parts := strings.Fields(string(out))
		if len(parts) > 0 {
			v := parts[len(parts)-1]
			tool.InstalledVersion = strings.TrimPrefix(v, "v")
			return
		}
	}
	// Binary exists but version unreadable — use manifest version as fallback.
	if tool.LatestVersion != "" {
		tool.InstalledVersion = tool.LatestVersion
	} else {
		tool.InstalledVersion = "?"
	}
}
