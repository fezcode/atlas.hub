# atlas.hub

![Banner Image](./banner-image.png)

**atlas.hub** is the central installer and manager for the **Atlas Suite**. It provides a unified, interactive TUI to discover, build, and install all Atlas tools from a single interface.

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey)

## ✨ Features

- 📦 **One-Click Installation:** Select multiple tools and install them in batch.
- 🛠️ **Automated Building:** Handles cloning and building via `gobake` automatically.
- 🎨 **Onyx & Gold Aesthetic:** Matches the premium look of the entire suite.
- 🔄 **Self-Contained:** Manages dependencies and installation paths (`~/.atlas/bin`).

## 🚀 Installation

To bootstrap the Atlas Suite, install `atlas.hub` first:

```bash
git clone https://github.com/fezcode/atlas.hub
cd atlas.hub
go run main.go
# Or build it:
gobake build
./build/atlas.hub-windows-amd64.exe
```

## ⌨️ Usage

Run the hub to see the available tools:
```bash
atlas.hub
```

1. **Select** tools using `Space`.
2. **Confirm** with `Enter`.
3. **Wait** for the installation to complete.
4. **Enjoy** your new tools in `~/.atlas/bin`.

## 📄 License
MIT License - see [LICENSE](LICENSE) for details.
