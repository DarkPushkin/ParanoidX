# Build Isle App APK with ParanoidX (V2Ray+Tor embedded)
# Run on Lenovo laptop (Windows 11) with Android Studio installed

$VERSION = "C21"
$PROJECT_DIR = "C:\Users\tomas\simplex-node\apps\isle_app"

Write-Host "=== Building Isle App APK v$VERSION ===" -ForegroundColor Cyan

# 1. Unset proxy (ParanoidX Tor proxy blocks Flutter pub)
$env:HTTP_PROXY = ""
$env:HTTPS_PROXY = ""
$env:ALL_PROXY = ""
$env:http_proxy = ""
$env:https_proxy = ""
$env:all_proxy = ""

# 2. Get dependencies
Write-Host "[1/4] Flutter pub get..." -ForegroundColor Yellow
cd $PROJECT_DIR
flutter pub get
if ($LASTEXITCODE -ne 0) { Write-Host "FAILED" -ForegroundColor Red; exit 1 }

# 3. Build APK (arm64 only for size)
Write-Host "[2/4] Building APK (arm64-v8a)..." -ForegroundColor Yellow
flutter build apk --debug --target-platform android-arm64 --split-per-abi
if ($LASTEXITCODE -ne 0) { Write-Host "FAILED" -ForegroundColor Red; exit 1 }

# 4. Verify APK
$APK = "$PROJECT_DIR\build\app\outputs\flutter-apk\app-arm64-v8a-debug.apk"
if (Test-Path $APK) {
    $size = (Get-Item $APK).Length / 1MB
    Write-Host "[3/4] APK: $APK ($([math]::Round($size, 1)) MB)" -ForegroundColor Green
} else {
    # Try release build
    flutter build apk --release --target-platform android-arm64 --split-per-abi
    $APK = "$PROJECT_DIR\build\app\outputs\flutter-apk\app-arm64-v8a-release.apk"
    if (Test-Path $APK) {
        $size = (Get-Item $APK).Length / 1MB
        Write-Host "[3/4] APK: $APK ($([math]::Round($size, 1)) MB)" -ForegroundColor Green
    } else {
        Write-Host "APK not found!" -ForegroundColor Red
        exit 1
    }
}

# 5. Install on connected device
Write-Host "[4/4] Installing on device..." -ForegroundColor Yellow
adb install -r $APK
if ($LASTEXITCODE -eq 0) {
    Write-Host "=== DONE === APK installed on device" -ForegroundColor Green
} else {
    Write-Host "=== DONE === APK built but install skipped (no device?)" -ForegroundColor Yellow
}
