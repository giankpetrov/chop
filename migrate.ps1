$ErrorActionPreference = "Stop"

$OldDir = "$env:USERPROFILE\bin"
$NewDir = "$env:LOCALAPPDATA\Programs\chop"
$Binary = "chop.exe"

Write-Host "chop migration: $OldDir -> $NewDir"
Write-Host ""

$OldPath = Join-Path $OldDir $Binary
if (-not (Test-Path $OldPath)) {
    Write-Host "chop not found in $OldDir — nothing to migrate."
    exit 0
}

# Move binary
New-Item -ItemType Directory -Force -Path $NewDir | Out-Null
Move-Item -Path $OldPath -Destination (Join-Path $NewDir $Binary) -Force
Write-Host "moved: $OldPath -> $(Join-Path $NewDir $Binary)"

# Update user PATH
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
$CleanOldDir = $OldDir.TrimEnd("\")
$CleanNewDir = $NewDir.TrimEnd("\")

# Remove old dir
$Parts = $UserPath -split ";" | Where-Object { $_.TrimEnd("\") -ne $CleanOldDir -and $_ -ne "" }

# Add new dir if not present
$HasNewDir = $Parts | ForEach-Object { $_.TrimEnd("\") } | Where-Object { $_ -eq $CleanNewDir }
if (-not $HasNewDir) {
    $Parts = @($NewDir) + $Parts
    Write-Host "added $NewDir to PATH"
}

if (($UserPath -split ";" | ForEach-Object { $_.TrimEnd("\") }) -contains $CleanOldDir) {
    Write-Host "removed $OldDir from PATH"
}

[Environment]::SetEnvironmentVariable("PATH", ($Parts -join ";"), "User")

# Update current session PATH
$env:PATH = ($env:PATH -split ";" | Where-Object { $_.TrimEnd("\") -ne $CleanOldDir } | ForEach-Object { $_ }) -join ";"
$CurrentPathParts = $env:PATH -split ";" | ForEach-Object { $_.TrimEnd("\") }
if ($CurrentPathParts -notcontains $CleanNewDir) {
    $env:PATH = "$NewDir;$env:PATH"
}

Write-Host ""
Write-Host "migration complete."
