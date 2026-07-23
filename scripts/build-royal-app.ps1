# Build Royal Flutter Client on Lenovo (Windows)
# Run from PowerShell in simplex-node directory

$ErrorActionPreference = "Stop"
$ROOT = Split-Path -Parent $PSScriptRoot
Set-Location "$ROOT/apps/royal_app"

Write-Host "=== Royal App Build (C1-C20) ===" -ForegroundColor Cyan

# Unset proxy for flutter (Tor proxy breaks pub get)
$env:HTTP_PROXY = ""
$env:HTTPS_PROXY = ""
$env:http_proxy = ""
$env:https_proxy = ""
$env:ALL_PROXY = ""
$env:all_proxy = ""

# Option A: Build for Web (served via browser)
Write-Host "`n[1/3] Pub get..." -ForegroundColor Yellow
flutter pub get

Write-Host "`n[2/3] Analyze..." -ForegroundColor Yellow
flutter analyze

Write-Host "`n[3/3] Build web..." -ForegroundColor Yellow
flutter build web --no-sound-null-safety

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n=== Build successful ===" -ForegroundColor Green
    Write-Host "Output: build/web/" -ForegroundColor Green
} else {
    Write-Host "`n=== Build FAILED ===" -ForegroundColor Red
    exit 1
}
