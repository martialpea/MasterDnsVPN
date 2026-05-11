@echo off
chcp 65001 >nul 2>&1
setlocal enabledelayedexpansion

:: ==============================================================================
:: MasterDnsVPN — اسکریپت اجرای سریع (Windows)
:: ==============================================================================

set "SCRIPT_DIR=%~dp0"
set "BIN_DIR=%SCRIPT_DIR%bin\"
set "CONFIG=%SCRIPT_DIR%client_config.toml"
set "RESOLVERS=%SCRIPT_DIR%resolvers.txt"

echo ══════════════════════════════════════════════
echo   🔒 MasterDnsVPN Client Launcher (Windows)
echo ══════════════════════════════════════════════
echo.

:: ── پیدا کردن باینری ──────────────────────────────────────────────────────
set "BIN="
if exist "%BIN_DIR%masterdnsvpn-client.exe"       set "BIN=%BIN_DIR%masterdnsvpn-client.exe"
if exist "%BIN_DIR%masterdnsvpn-client-win.exe"   set "BIN=%BIN_DIR%masterdnsvpn-client-win.exe"
if exist "%SCRIPT_DIR%masterdnsvpn-client.exe"    set "BIN=%SCRIPT_DIR%masterdnsvpn-client.exe"

if "%BIN%"=="" (
    echo ❌ باینری کلاینت یافت نشد!
    echo.
    echo باینری را از Releases دانلود کنید:
    echo   https://github.com/masterking32/MasterDnsVPN/releases
    echo.
    echo سپس در پوشه bin\ قرار دهید.
    echo.
    pause
    exit /b 1
)

echo ✅ باینری: %BIN%

:: ── بررسی کانفیگ ──────────────────────────────────────────────────────────
if not exist "%CONFIG%" (
    echo ❌ فایل کانفیگ یافت نشد: %CONFIG%
    echo.
    echo راه‌حل:
    echo   ۱. فایل launcher.html را در مرورگر باز کنید
    echo   ۲. تنظیمات را وارد کنید
    echo   ۳. دکمه دانلود client_config.toml بزنید
    echo   ۴. فایل را در کنار این اسکریپت قرار دهید
    echo.
    pause
    exit /b 1
)

if not exist "%RESOLVERS%" (
    echo ⚠  resolvers.txt یافت نشد — از 8.8.8.8 استفاده می‌شود
    echo 8.8.8.8> "%RESOLVERS%"
)

echo ✅ کانفیگ: %CONFIG%
echo ✅ Resolvers: %RESOLVERS%
echo.
echo 🚀 در حال راه‌اندازی...
echo    برای خروج: Ctrl+C
echo ══════════════════════════════════════════════
echo.

"%BIN%" -config "%CONFIG%" -resolvers "%RESOLVERS%" %*

if errorlevel 1 (
    echo.
    echo ❌ برنامه با خطا خارج شد.
    pause
)
