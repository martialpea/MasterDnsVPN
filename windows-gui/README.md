# 🪟 MasterDnsVPN — Windows GUI

رابط گرافیکی بومی ویندوز برای MasterDnsVPN.

## نحوه کار

این برنامه یک **EXE کوچک** است که:
1. یک HTTP server روی `localhost` با یک port تصادفی راه‌اندازی می‌کند
2. پنجره مرورگر پیش‌فرض سیستم را باز می‌کند (از WebView2 داخلی ویندوز استفاده می‌کند)
3. باینری اصلی (`masterdnsvpn-client.exe`) را به عنوان یک پروسه فرزند اجرا می‌کند
4. لاگ‌های باینری را real-time در رابط گرافیکی نشان می‌دهد

## ساختار پوشه روی ویندوز

```
MasterDnsVPN-Windows/
├── MasterDnsVPN-GUI.exe          ← این را اجرا کنید ✅
├── masterdnsvpn-client.exe       ← باینری اصلی (از Releases)
├── client_config.toml            ← خودکار ساخته می‌شود
└── resolvers.txt                 ← خودکار ساخته می‌شود
```

## نحوه Build

### روش A — ویندوز (با PowerShell)
```powershell
cd windows-gui
.\build-windows.ps1
```

### روش B — لینوکس/macOS (Cross-compile)
```bash
cd windows-gui
chmod +x build-crosscompile.sh
./build-crosscompile.sh
```

### روش C — GitHub Actions (خودکار)
فایل `.github/workflows/build-windows-gui.yml` موجود است.
با هر push روی branch اصلی، یک artifact آماده می‌شود.

## پیش‌نیازها

- Go 1.21+
- ویندوز 10/11 (برای اجرا)
- مرورگر پیش‌فرض سیستم (Chrome/Edge/Firefox)

## قابلیت‌ها

- ✅ اجرای واقعی باینری (نه شبیه‌سازی)
- ✅ لاگ‌های real-time از باینری اصلی
- ✅ ساخت خودکار `client_config.toml` از فرم
- ✅ قطع و وصل کردن بدون نیاز به Command Prompt
- ✅ نمایش PID و وضعیت اتصال
- ✅ بدون نیاز به Node.js یا Electron
- ✅ حجم کم (زیر ۵ مگابایت)
