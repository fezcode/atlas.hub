# 1. Setup Atlas Bin Directory
$AtlasDir = Join-Path $HOME ".atlas\bin"
if (!(Test-Path $AtlasDir)) { New-Item -ItemType Directory -Path $AtlasDir }

# 2. Check for Go
if (!(Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Error: Go is not installed. Please install Go (1.25+) first." -ForegroundColor Red
    exit 1
}

# 3. Check for gobake
if (!(Get-Command gobake -ErrorAction SilentlyContinue)) {
    Write-Host "🛰️  Installing gobake..." -ForegroundColor Cyan
    go install github.com/fezcode/gobake/cmd/gobake@latest
    $GoBin = go env GOBIN
    if ($null -eq $GoBin -or $GoBin -eq "") {
        $GoPath = go env GOPATH
        $GoBin = Join-Path $GoPath "bin"
    }
    $env:PATH = "$GoBin;$env:PATH"
    
    if (!(Get-Command gobake -ErrorAction SilentlyContinue)) {
        Write-Host "❌ Error: gobake could not be installed/found." -ForegroundColor Red
        exit 1
    }
}

# 4. Clone and Build atlas.hub
$TempDir = Join-Path $env:TEMP "atlas-hub-bootstrap"
if (Test-Path $TempDir) { Remove-Item -Recurse -Force $TempDir }
New-Item -ItemType Directory -Path $TempDir

Write-Host "🛰️  Bootstrapping atlas.hub..." -ForegroundColor Cyan
Set-Location $TempDir
git clone https://github.com/fezcode/atlas.hub.git .

Write-Host "🛰️  Building..." -ForegroundColor Cyan
gobake build

# 5. Detect System Info
$OS = "windows"
$Arch = $env:PROCESSOR_ARCHITECTURE.ToLower()
if ($Arch -eq "amd64") { $Arch = "amd64" }
elseif ($Arch -eq "arm64") { $Arch = "arm64" }

$Binary = "build\atlas.hub-$OS-$Arch.exe"

# 6. Relocate
if (Test-Path $Binary) {
    Move-Item -Path $Binary -Destination "$AtlasDir\atlas.hub.exe" -Force
    Write-Host "✅ atlas.hub installed to $AtlasDir\atlas.hub.exe" -ForegroundColor Green
} else {
    Write-Host "❌ Error: Build failed, binary not found." -ForegroundColor Red
    exit 1
}

# 7. Cleanup
Set-Location $HOME
Remove-Item -Recurse -Force $TempDir

Write-Host "🚀 Starting Atlas Hub..." -ForegroundColor Green
& "$AtlasDir\atlas.hub.exe"
