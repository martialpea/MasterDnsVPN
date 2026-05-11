#!/bin/bash
# ==============================================================================
# MasterDnsVPN — ساخت GUI ویندوز (از Linux/macOS)
# ==============================================================================
# پیش‌نیاز: Go 1.21+ با cross-compile support
# اجرا از پوشه windows-gui:
#   chmod +x build-crosscompile.sh
#   ./build-crosscompile.sh
# ==============================================================================

set -e

echo "══════════════════════════════════════════"
echo "  MasterDnsVPN — Windows GUI Cross-Compile"
echo "══════════════════════════════════════════"
echo ""

# بررسی Go
if ! command -v go &>/dev/null; then
    echo "[ERR] Go یافت نشد! از https://golang.org/dl/ نصب کنید"
    exit 1
fi
echo "[OK] Go: $(go version)"

# Cross-compile برای ویندوز
echo "[>>] در حال ساخت MasterDnsVPN-GUI.exe ..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
  go build \
    -ldflags="-H windowsgui -s -w" \
    -o ../MasterDnsVPN-GUI.exe \
    .

echo "[OK] MasterDnsVPN-GUI.exe آماده است!"
echo ""
echo "قدم بعدی:"
echo "  ۱. MasterDnsVPN-GUI.exe را کنار masterdnsvpn-client.exe روی ویندوز قرار دهید"
echo "  ۲. MasterDnsVPN-GUI.exe را اجرا کنید"
