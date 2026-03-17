$ErrorActionPreference = "Stop"

$Repo = "AgusRdz/chop"
$InstallDir = if ($env:CHOP_INSTALL_DIR) { $env:CHOP_INSTALL_DIR } else { "$env:LOCALAPPDATA\Programs\chop" }

# Detect architecture
$Arch = if ([System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture -eq [System.Runtime.InteropServices.Architecture]::Arm64) {
    "arm64"
} else {
    "amd64"
}

$Binary = "chop-windows-$Arch.exe"

# Get latest version
if (-not $env:CHOP_VERSION) {
    $Release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
    $env:CHOP_VERSION = $Release.tag_name
}

if (-not $env:CHOP_VERSION) {
    Write-Error "failed to determine latest version"
    exit 1
}

$Url = "https://github.com/$Repo/releases/download/$($env:CHOP_VERSION)/$Binary"

Write-Host "installing chop $($env:CHOP_VERSION) (windows/$Arch)..."

# Create install dir
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

# Download binary
$Destination = Join-Path $InstallDir "chop.exe"
Invoke-WebRequest -Uri $Url -OutFile $Destination

Write-Host "installed chop to $Destination"
Write-Host ""

# Add to user PATH if not already present
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
$CleanInstallDir = $InstallDir.TrimEnd("\")
$PathParts = $UserPath -split ";" | ForEach-Object { $_.TrimEnd("\") }

if ($PathParts -notcontains $CleanInstallDir) {
    [Environment]::SetEnvironmentVariable("PATH", "$InstallDir;$UserPath", "User")
    Write-Host "added $InstallDir to PATH"
}

# Update current session PATH so it can be used immediately
$CurrentPathParts = $env:PATH -split ";" | ForEach-Object { $_.TrimEnd("\") }
if ($CurrentPathParts -notcontains $CleanInstallDir) {
    $env:PATH = "$InstallDir;$env:PATH"
}

Write-Host "next steps:"
Write-Host ""
Write-Host "  # use directly with any command:"
Write-Host "  chop git status"
Write-Host "  chop docker ps"
Write-Host ""
Write-Host "  # claude code hook (auto-rewrite bash tool calls):"
Write-Host "  chop init --global"
Write-Host "  chop init --status"
