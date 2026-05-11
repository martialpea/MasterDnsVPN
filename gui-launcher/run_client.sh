#!/usr/bin/env bash
# ==============================================================================
# MasterDnsVPN — اسکریپت اجرای سریع (Linux / macOS)
# ==============================================================================
# استفاده:
#   chmod +x run_client.sh
#   ./run_client.sh
#   ./run_client.sh -config my_config.toml
# ==============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="$SCRIPT_DIR/bin"
CONFIG="$SCRIPT_DIR/client_config.toml"
RESOLVERS="$SCRIPT_DIR/resolvers.txt"

# ── پیدا کردن باینری ─────────────────────────────────────────────────────────
find_binary() {
    local candidates=(
        "$BIN_DIR/masterdnsvpn-client"
        "$BIN_DIR/masterdnsvpn-client-linux-amd64"
        "$BIN_DIR/masterdnsvpn-client-darwin-amd64"
        "$BIN_DIR/masterdnsvpn-client-darwin-arm64"
        "$SCRIPT_DIR/masterdnsvpn-client"
        "$(which masterdnsvpn-client 2>/dev/null || echo '')"
    )
    for c in "${candidates[@]}"; do
        if [[ -n "$c" && -x "$c" ]]; then
            echo "$c"; return 0
        fi
    done
    return 1
}

# ── بررسی وجود کانفیگ ────────────────────────────────────────────────────────
check_config() {
    if [[ ! -f "$CONFIG" ]]; then
        echo "❌ فایل کانفیگ یافت نشد: $CONFIG"
        echo ""
        echo "راه‌حل:"
        echo "  ۱. فایل launcher.html را در مرورگر باز کنید"
        echo "  ۲. تنظیمات را وارد کنید"
        echo "  ۳. دکمه 'دانلود client_config.toml' بزنید"
        echo "  ۴. فایل را در کنار این اسکریپت قرار دهید"
        exit 1
    fi

    if [[ ! -f "$RESOLVERS" ]]; then
        echo "⚠ فایل resolvers.txt یافت نشد — از 8.8.8.8 استفاده می‌شود"
        echo "8.8.8.8" > "$RESOLVERS"
    fi
}

# ── اجرای اصلی ───────────────────────────────────────────────────────────────
main() {
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  🔒 MasterDnsVPN Client Launcher"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    local bin
    if ! bin=$(find_binary); then
        echo "❌ باینری کلاینت یافت نشد!"
        echo ""
        echo "باینری را از Releases دانلود کنید:"
        echo "  https://github.com/masterking32/MasterDnsVPN/releases"
        echo ""
        echo "سپس در پوشه bin/ قرار دهید:"
        echo "  $BIN_DIR/"
        exit 1
    fi

    echo "✅ باینری: $bin"

    check_config

    echo "✅ کانفیگ: $CONFIG"
    echo "✅ Resolvers: $RESOLVERS"
    echo ""
    echo "🚀 در حال راه‌اندازی..."
    echo "   برای خروج: Ctrl+C"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    # اجرا با forwarding سیگنال
    exec "$bin" -config "$CONFIG" -resolvers "$RESOLVERS" "$@"
}

main "$@"
