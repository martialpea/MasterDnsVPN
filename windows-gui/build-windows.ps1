# ==============================================================================
# MasterDnsVPN — ساخت GUI ویندوز
# نویسنده: MasterkinG32
# ==============================================================================
# پیش‌نیاز: Go 1.21+ نصب باشد
# اجرا:
#   cd windows-gui
#   .\build-windows.ps1
# ==============================================================================

$ErrorActionPreference = "Stop"

Write-Host "══════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "  MasterDnsVPN — Windows GUI Builder" -ForegroundColor Cyan
Write-Host "══════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""

# Check Go
try {
    $goVer = go version
    Write-Host "[OK] Go: $goVer" -ForegroundColor Green
} catch {
    Write-Host "[ERR] Go یافت نشد! از https://golang.org/dl/ نصب کنید" -ForegroundColor Red
    exit 1
}

# Build GUI
Write-Host "[>>] در حال ساخت MasterDnsVPN-GUI.exe ..." -ForegroundColor Yellow
$env:GOOS   = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

go build -ldflags="-H windowsgui -s -w" -o "..\MasterDnsVPN-GUI.exe" "."

if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERR] Build شکست خورد!" -ForegroundColor Red
    exit 1
}

Write-Host "[OK] MasterDnsVPN-GUI.exe ساخته شد!" -ForegroundColor Green
Write-Host ""
Write-Host "قدم بعدی:" -ForegroundColor Cyan
Write-Host "  ۱. MasterDnsVPN-GUI.exe را کنار masterdnsvpn-client.exe قرار دهید"
Write-Host "  ۲. MasterDnsVPN-GUI.exe را اجرا کنید"
Write-Host "  ۳. دامنه و کلید سرور را وارد کنید و دکمه اتصال را بزنید"
Write-Host ""
