$ErrorActionPreference = "Stop"

$Repo = "giankpetrov/openchop"
$InstallDir = if ($env:OPENCHOP_INSTALL_DIR) { $env:OPENCHOP_INSTALL_DIR } else { "$env:LOCALAPPDATA\Programs\openchop" }

# Detect architecture
$Arch = if ([System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture -eq [System.Runtime.InteropServices.Architecture]::Arm64) {
    "arm64"
} else {
    "amd64"
}

$Binary = "openchop-windows-$Arch.exe"

# Get latest version
if (-not $env:OPENCHOP_VERSION) {
    $Release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
    $env:OPENCHOP_VERSION = $Release.tag_name
}

if (-not $env:OPENCHOP_VERSION) {
    Write-Error "failed to determine latest version"
    exit 1
}

$Url = "https://github.com/$Repo/releases/download/$($env:OPENCHOP_VERSION)/$Binary"

Write-Host "installing openchop $($env:OPENCHOP_VERSION) (windows/$Arch)..."

# Create install dir
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

# Download binary
$Destination = Join-Path $InstallDir "openchop.exe"
Invoke-WebRequest -Uri $Url -OutFile $Destination

Write-Host "installed openchop to $Destination"
Write-Host ""

# Add to user PATH if not already present
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$InstallDir;$UserPath", "User")
    Write-Host "added $InstallDir to PATH"
    Write-Host "restart your terminal for PATH changes to take effect"
    Write-Host ""
}

Write-Host "next steps:"
Write-Host ""
Write-Host "  # use directly with any command:"
Write-Host "  openchop git status"
Write-Host "  openchop docker ps"
Write-Host ""
Write-Host "  # claude code hook (auto-rewrite bash tool calls):"
Write-Host "  openchop init --global"
Write-Host "  openchop init --status"
