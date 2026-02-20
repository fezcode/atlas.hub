# Atlas Suite Development Guidelines

When creating or maintaining an Atlas Project (e.g., `atlas.xyz`), strict adherence to the following standards is required.

## 1. Build System: gobake
All projects must use **gobake** for orchestration.
- **Recipe.go:** Must exist in the root.
- **recipe.piml:** Must exist in the root and define metadata (`name`, `version`, `description`, `license`).

## 2. Versioning & Embedding
Versions are managed in `recipe.piml` and injected via `-ldflags` during build.

**recipe.piml:**
```piml
(name) atlas.tool
(version) 0.1.0
...
```

**Recipe.go:**
You must manually inject the version using `ldflags` in the build task. **Do NOT** use `ctx.BakeBinary` directly if it doesn't support ldflags injection (as of gobake v0.3.0). Use `ctx.Run` instead:

```go
ldflags := fmt.Sprintf("-X main.Version=%s", bake.Info.Version)
err := ctx.Run("go", "build", "-ldflags", ldflags, "-o", output, ".")
```

**main.go:**
Must declare a version variable and handle the flag:
```go
var Version = "dev"

func main() {
    if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
        fmt.Printf("atlas.tool v%s
", Version)
        return
    }
    // ...
}
```

## 3. Central Registry (atlas.hub)
Every new tool or version update **MUST** be reflected in `atlas.hub`.

1.  **Update Manifest:** Edit `atlas.hub/manifest.piml` to include the new tool or update the version of an existing one.
    ```piml
    > (tool)
      (name) atlas.newtool
      (description) ...
      (version) 0.1.0
      ...
    ```
2.  **Push Changes:** Commit and push the changes to `atlas.hub` immediately so users receive the update via the live sync mechanism.
