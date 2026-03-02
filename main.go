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
	"github.com/go-git/go-git/v5"
)

//go:embed manifest.piml
var manifestFS embed.FS

var Version = "0.1.0"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("atlas.hub v%s\n", Version)
		return
	}

	// 1. Determine Paths
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home dir: %v\n", err)
		os.Exit(1)
	}
	atlasDir := filepath.Join(home, ".atlas")
	installPath := filepath.Join(atlasDir, "bin")
	hubDataDir := filepath.Join(atlasDir, "hub-data")

	// 2. Sync Manifest from Repo (Live Updates)
	fmt.Print("🛰️  Syncing Atlas manifest...")
	manifestPath, err := syncManifest(hubDataDir)
	var data []byte
	if err != nil {
		fmt.Printf(" (offline, using embedded)\n")
		data, _ = manifestFS.ReadFile("manifest.piml")
	} else {
		fmt.Printf(" done\n")
		data, err = os.ReadFile(manifestPath)
		if err != nil {
			data, _ = manifestFS.ReadFile("manifest.piml")
		}
	}

	// 3. Parse Manifest
	var manifest model.Manifest
	if err := piml.Unmarshal(data, &manifest); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing manifest: %v\n", err)
		os.Exit(1)
	}

	tools := manifest.Tools
	if len(tools) == 0 {
		fmt.Fprintf(os.Stderr, "No tools found in manifest.\n")
		os.Exit(1)
	}

	// 4. Init Manager
	manager, err := install.NewManager(installPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing manager: %v\n", err)
		os.Exit(1)
	}
	defer manager.Cleanup()

	// Check installed versions
	for i := range tools {
		manager.CheckInstalledVersion(&tools[i])
	}

	// 5. Start TUI
	p := tea.NewProgram(ui.NewModel(manager, tools, installPath), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

func syncManifest(hubDataDir string) (string, error) {
	repoURL := "https://github.com/fezcode/atlas.hub.git"
	manifestFile := filepath.Join(hubDataDir, "manifest.piml")

	if _, err := os.Stat(hubDataDir); os.IsNotExist(err) {
		// Clone for the first time
		_, err = git.PlainClone(hubDataDir, false, &git.CloneOptions{
			URL: repoURL,
		})
		if err != nil {
			return "", err
		}
	} else {
		// Open and Pull
		r, err := git.PlainOpen(hubDataDir)
		if err != nil {
			return "", err
		}
		w, err := r.Worktree()
		if err != nil {
			return "", err
		}
		
		// Force clean the working directory to prevent uncommitted changes from blocking the pull
		err = w.Clean(&git.CleanOptions{Dir: true})
		if err != nil {
			// Log error but proceed
			fmt.Fprintf(os.Stderr, " warning: failed to clean hub-data: %v", err)
		}
		
		err = w.Checkout(&git.CheckoutOptions{Force: true})
		if err != nil {
			// Log error but proceed
			fmt.Fprintf(os.Stderr, " warning: failed to force checkout hub-data: %v", err)
		}

		err = w.Pull(&git.PullOptions{RemoteName: "origin", Force: true})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return "", err
		}
	}

	return manifestFile, nil
}
