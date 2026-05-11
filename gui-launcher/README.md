# 🔒 MasterDnsVPN Launcher

## محتویات این پکیج

```
MasterDnsVPN-Launcher/
├── launcher.html          ← GUI مرورگری (اینجا شروع کنید!)
├── client_config.toml     ← کانفیگ نمونه (ویرایش کنید)
├── resolvers.txt          ← لیست DNS resolver
├── run_client.sh          ← اجراکننده Linux/macOS
├── run_client.bat         ← اجراکننده Windows
├── README.md
└── bin/                   ← باینری کامپایل‌شده را اینجا بگذارید
    └── masterdnsvpn-client  (دانلود از GitHub Releases)
```

---

## ۳ مرحله برای اتصال

### مرحله ۱ — دانلود باینری

از صفحه Releases پروژه باینری مناسب سیستم خود را دانلود کنید:

```
https://github.com/masterking32/MasterDnsVPN/releases
```

فایل را در پوشه `bin/` قرار دهید:
- **Linux x64:** `bin/masterdnsvpn-client`
- **macOS:** `bin/masterdnsvpn-client`
- **Windows:** `bin/masterdnsvpn-client.exe`

---

### مرحله ۲ — تنظیم کانفیگ

**روش ۱ — GUI (آسان):**
1. فایل `launcher.html` را در مرورگر باز کنید
2. دامنه و کلید سرور را وارد کنید
3. دکمه **دانلود client_config.toml** را بزنید
4. فایل دانلودشده را جایگزین `client_config.toml` کنید

**روش ۲ — دستی:**
فایل `client_config.toml` را باز کرده و این موارد را تنظیم کنید:

```toml
DOMAINS = ["v.yourdomain.com"]   # دامنه تونل سرور
ENCRYPTION_KEY = "your-key"       # کلید مشترک با سرور
DATA_ENCRYPTION_METHOD = 5        # باید با سرور یکسان باشد
LISTEN_PORT = 18000               # پورت SOCKS5 محلی
```

---

### مرحله ۳ — اجرا

**Linux/macOS:**
```bash
chmod +x run_client.sh
./run_client.sh
```

**Windows:**
```
run_client.bat را دو بار کلیک کنید
```

**دستی:**
```bash
./bin/masterdnsvpn-client -config client_config.toml -resolvers resolvers.txt
```

---

### تنظیم پراکسی

بعد از اتصال، برنامه‌های خود را روی SOCKS5 تنظیم کنید:

| تنظیم | مقدار |
|--------|-------|
| Protocol | SOCKS5 |
| Host | 127.0.0.1 |
| Port | 18000 |

**مرورگر Firefox:**
Settings → Network → Manual proxy → SOCKS Host: 127.0.0.1 Port: 18000

**curl:**
```bash
curl -x socks5://127.0.0.1:18000 https://example.com
```

---

## عیب‌یابی

| مشکل | راه‌حل |
|-------|---------|
| `❌ باینری یافت نشد` | باینری را در پوشه `bin/` قرار دهید |
| `❌ Session Init failed` | دامنه یا کلید با سرور مطابقت ندارد |
| `MTU test failed (all resolvers)` | Resolver ها فیلتر هستند — تغییر دهید |
| اتصال برقرار اما سایت باز نمی‌شود | پراکسی مرورگر را بررسی کنید |

---

## GUI چگونه کار می‌کند؟

`launcher.html` یک رابط گرافیکی مرورگری است که:
- کانفیگ را به‌صورت خودکار می‌سازد
- خروجی شبیه‌سازی‌شده برنامه را نشان می‌دهد
- فایل‌های کانفیگ را آماده دانلود می‌کند

برای GUI کامل با اجرای واقعی باینری، پروژه باید با **Wails** یا **Electron** بسته‌بندی شود.
