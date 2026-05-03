# verify-release.ps1
# Advanced Stability & Portability Audit for Docklog (Windows)

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "--- [Docklog Stability Audit: Windows] ---" -ForegroundColor Cyan

# 1. Resolve Environment
Write-Host "[1/5] Checking environment..." -NoNewline
$goPath = go env GOPATH
if (-not $goPath) {
    Write-Host " FAILED" -ForegroundColor Red
    exit 1
}
if (-not (Test-Path "dist")) { New-Item -ItemType Directory -Path "dist" -Force | Out-Null }
Write-Host " DONE" -ForegroundColor Green

# 2. Build Audit (Static Linking)
Write-Host "[2/5] Building static binary (CGO_ENABLED=0)..." -NoNewline
$env:CGO_ENABLED = "0"
try {
    go build -trimpath -o ./dist/docklog-audit-win.exe .
    Write-Host " DONE" -ForegroundColor Green
} catch {
    Write-Host " FAILED" -ForegroundColor Red
    exit 1
}

# 3. Cross-Compile Verification
Write-Host "[3/5] Verifying multi-architecture builds..." -NoNewline
try {
    $env:GOOS = "linux"; $env:GOARCH = "amd64"; go build -o ./dist/tmp_linux .
    $env:GOOS = "linux"; $env:GOARCH = "arm64"; go build -o ./dist/tmp_arm .
    # Reset env
    $env:GOOS = ""; $env:GOARCH = ""
    Write-Host " DONE" -ForegroundColor Green
} catch {
    Write-Host " FAILED" -ForegroundColor Red
    exit 1
}

# 4. Binary Integrity Check
Write-Host "[4/5] Running Binary Integrity Check..." -NoNewline
try {
    $versionOutput = & ./dist/docklog-audit-win.exe version
    if ($versionOutput -like "*Docklog*") {
        Write-Host " SUCCESS" -ForegroundColor Green
    } else {
        throw "Unexpected output"
    }
} catch {
    Write-Host " FAILED" -ForegroundColor Red
    Write-Warning "Binary failed to start or produced invalid output."
    exit 1
}

# 5. Final Installation
Write-Host "[5/5] Performing final local installation..." -NoNewline
Stop-Process -Name docklog -ErrorAction SilentlyContinue 2>$null
go install .
Write-Host " DONE" -ForegroundColor Green

Write-Host ""
Write-Host "Installed Version:" -ForegroundColor Cyan
$finalBin = Join-Path $goPath "bin\docklog.exe"
& $finalBin version

Write-Host ""
Write-Host "RELEASE STABILITY AUDIT PASSED" -ForegroundColor Green
Write-Host "This version is safe to publish." -ForegroundColor Cyan
Write-Host ""
