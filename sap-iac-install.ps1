# Stop on error, but handle exceptions gracefully via try-catch
$ErrorActionPreference = "Stop"

$InstallDir = "$env:LOCALAPPDATA\Programs\SAP-IAC"
$ExePath = "$InstallDir\sap-iac.exe"
$Url = "https://github.com/SAP/iac-spec-kit-for-sap-business-ai-platform/releases/download/v0.1.0/sap-iac_0.1.0_windows_amd64.exe"

Write-Host "Installing SAP IAC..." -ForegroundColor Cyan

# 1. Ensure target directory exists
try {
    if (-not (Test-Path -Path $InstallDir)) {
        New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    }
} catch {
    Write-Host "[ERROR] Failed to create installation folder '$InstallDir':$_" -ForegroundColor Red
    exit 1
}

# 2. Download binary with HTTP status check
try {
    Write-Host "Downloading binary from GitHub..." -ForegroundColor Yellow
    Invoke-WebRequest -Uri $Url -OutFile$ExePath -UserAgent "PowerShell-Installer"
} catch {
    Write-Host "[ERROR] Download failed!" -ForegroundColor Red
    Write-Host "Details: $_" -ForegroundColor Red
    Write-Host "Please check your internet connection or verify the release URL." -ForegroundColor Yellow
    exit 1
}

# 3. Validate downloaded file (Ensure it is not empty or an HTML error page)
if (-not (Test-Path -Path $ExePath) -or (Get-Item$ExePath).Length -lt 1MB) {
    Write-Host "[ERROR] Downloaded file appears corrupted or invalid (Size < 1MB)." -ForegroundColor Red
    Remove-Item -Path $ExePath -Force -ErrorAction SilentlyContinue
    exit 1
}

# 4. Update persistent User PATH variable
try {
    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($UserPath -notlike "*$InstallDir*") {
        $NewPath = if ([string]::IsNullOrWhiteSpace($UserPath)) {$InstallDir } else { "$UserPath;$InstallDir" }
        [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
    }
} catch {
    Write-Host "[WARNING] Failed to set permanent User PATH environment variable: $_" -ForegroundColor Yellow
}

# 5. Refresh current session PATH so 'sap-iac' works immediately
if ($env:Path -notlike "*$InstallDir*") {
    $env:Path += ";$InstallDir"
}

# 6. Verification test
try {
    $versionOutput = & "$ExePath" --version 2>&1
    Write-Host "`n[SUCCESS] SAP IAC installed successfully!" -ForegroundColor Green
    Write-Host "Location: $ExePath" -ForegroundColor Gray
    Write-Host "Test Run Output: $versionOutput" -ForegroundColor Gray
    Write-Host "`nYou can now run 'sap-iac' from any terminal session." -ForegroundColor Green
} catch {
    Write-Host "[WARNING] Binary installed to '$ExePath', but failed to execute test command." -ForegroundColor Yellow
}
