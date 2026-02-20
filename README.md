# atlas.hub 🛰️

![Banner Image](./banner-image.png)

**atlas.hub** is the central entry point and interactive installer for the **Atlas Suite**. Instead of manually managing a dozen CLI tools, the hub provides a unified interface to discover, build, and maintain your Atlas toolkit with a high-fidelity "Onyx & Gold" aesthetic.

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey)

## 🚀 One-Liner Installation

Run the appropriate command for your operating system to bootstrap the Atlas Suite. The script will verify your environment, install `atlas.hub`, and start it immediately.

### 🐧 Linux / 🍎 macOS (Bash)
```bash
curl -fsSL https://raw.githubusercontent.com/fezcode/atlas.hub/main/scripts/install.sh | bash
```

### 🪟 Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/fezcode/atlas.hub/main/scripts/install.ps1 | iex
```

---

## 📋 Prerequisites

Before running the installer, ensure you have the following installed and available in your `PATH`:

1.  **Go (1.25+):** Required for building the tools from source. [Download Go](https://go.dev/dl/).
2.  **gobake:** The Atlas build orchestrator. If not found, the installer will attempt to install it for you via `go install github.com/fezcode/gobake@latest`.

---

## ✨ Features

- 📦 **Interactive TUI:** A beautiful checklist to select and batch-install tools.
- 🔄 **Live Sync:** Automatically pulls the latest `manifest.piml` from GitHub on every run.
- 🛠️ **Automated Builds:** Clones and builds every selected tool using `gobake`.
- 📂 **Centralized Bin:** Installs everything to a clean `~/.atlas/bin` directory.
- 📶 **Offline Fallback:** Uses an embedded manifest if you're not connected to the internet.

---

## ⚙️ Post-Installation: Add to PATH

To use your new Atlas tools from anywhere, you **must** add the installation directory to your system's `PATH`.

### 🪟 Windows
Run this command in an **Administrator PowerShell** to permanently update your User PATH:
```powershell
$Path = [Environment]::GetEnvironmentVariable("Path", "User")
$AtlasPath = "$HOME\.atlas\bin"
if ($Path -notlike "*$AtlasPath*") {
    [Environment]::SetEnvironmentVariable("Path", "$Path;$AtlasPath", "User")
    Write-Host "✅ Added to PATH. Restart your terminal to apply changes." -ForegroundColor Green
}
```

### 🍎 macOS (Zsh)
```bash
echo 'export PATH="$HOME/.atlas/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### 🐧 Linux (Bash)
```bash
echo 'export PATH="$HOME/.atlas/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

---

## 🛠️ Included Tools

The Atlas Hub currently manages the following tools:
- **atlas.todo**: Fast, minimalist task manager.
- **atlas.stats**: Real-time system monitoring.
- **atlas.websearch**: Interactive CLI search (Wiki, Reddit, HN).
- **atlas.compass**: Secure local-first password manager.
- **atlas.clock**: High-visibility world clock.
- **atlas.cam**: Terminal webcam viewer & ASCII camera.
- **atlas.games**: Shell-based game collection.
- **atlas.bench**: High-performance CLI benchmarking.
- **atlas.radar**: Git workspace monitor.
- **atlas.otp**: Secure TOTP (2FA) manager.
- **atlas.diff**: Side-by-side terminal diff tool.

---

## 🏗️ Manual Build
If you prefer to build manually:
```bash
git clone https://github.com/fezcode/atlas.hub
cd atlas.hub
gobake build
./build/atlas.hub-windows-amd64.exe  # Or your platform equivalent
```

## 📄 License
MIT License - see [LICENSE](LICENSE) for details.
