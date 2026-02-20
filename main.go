package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"atlas.hub/internal/install"
	"atlas.hub/internal/model"
	"atlas.hub/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fezcode/go-piml"
)

//go:embed manifest.piml
var manifestFS embed.FS

func main() {
	// Load manifest
	data, err := manifestFS.ReadFile("manifest.piml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading embedded manifest: %v\n", err)
		os.Exit(1)
	}

	var manifest model.Manifest
	if err := piml.Unmarshal(data, &manifest); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing manifest: %v\n", err)
		os.Exit(1)
	}

	if len(manifest.Tools) == 0 {
		var nested struct {
			ToolsSection struct {
				Tools []model.Tool `piml:"tool"`
			} `piml:"tools"`
		}
		if err := piml.Unmarshal(data, &nested); err == nil && len(nested.ToolsSection.Tools) > 0 {
			manifest.Tools = nested.ToolsSection.Tools
		}
	}

	// Determine install path
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home dir: %v\n", err)
		os.Exit(1)
	}
	installPath := filepath.Join(home, ".atlas", "bin")

	// Init manager
	manager, err := install.NewManager(installPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing manager: %v\n", err)
		os.Exit(1)
	}
	defer manager.Cleanup()

	// Start TUI
	p := tea.NewProgram(ui.NewModel(manager, manifest.Tools, installPath), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
