//go:build windows

// ==============================================================================
// MasterDnsVPN — Windows GUI Launcher
// Author: MasterkinG32
// Github: https://github.com/masterking32
// Year: 2026
//
// نحوه Build:
//   go build -ldflags="-H windowsgui" -o MasterDnsVPN-GUI.exe .
//
// این برنامه یک HTTP server داخلی روی localhost اجرا می‌کند،
// سپس پنجره مرورگر سیستم را باز می‌کند.
// باینری اصلی (masterdnsvpn-client.exe) را در کنار این EXE قرار دهید.
// ==============================================================================

package main

import (
	"bufio"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

//go:embed ui/index.html ui/assets
var uiFiles embed.FS

// ─── Process management ───────────────────────────────────────────────────────

type VPNProcess struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	cancel  context.CancelFunc
	running bool
	logs    []LogLine
	maxLogs int
}

type LogLine struct {
	Time  string `json:"time"`
	Level string `json:"level"`
	Msg   string `json:"msg"`
}

var vpn = &VPNProcess{maxLogs: 500}

func (v *VPNProcess) addLog(level, msg string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	line := LogLine{
		Time:  time.Now().Format("15:04:05"),
		Level: level,
		Msg:   msg,
	}
	v.logs = append(v.logs, line)
	if len(v.logs) > v.maxLogs {
		v.logs = v.logs[1:]
	}
}

func (v *VPNProcess) Start(binPath, configPath, resolversPath string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.running {
		return fmt.Errorf("already running")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, binPath,
		"-config", configPath,
		"-resolvers", resolversPath,
	)
	// Hide console window
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return err
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return err
	}

	v.cmd = cmd
	v.cancel = cancel
	v.running = true
	v.logs = nil

	// Read stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			level := classifyLine(line)
			v.addLog(level, stripANSI(line))
		}
	}()

	// Read stderr
	go func() {
		data, _ := io.ReadAll(stderr)
		if len(data) > 0 {
			for _, l := range strings.Split(string(data), "\n") {
				if strings.TrimSpace(l) != "" {
					v.addLog("error", stripANSI(l))
				}
			}
		}
	}()

	// Watch for exit
	go func() {
		_ = cmd.Wait()
		v.mu.Lock()
		v.running = false
		v.cancel = nil
		v.cmd = nil
		v.mu.Unlock()
		v.addLog("warn", "⚠ برنامه متوقف شد")
	}()

	return nil
}

func (v *VPNProcess) Stop() {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.cancel != nil {
		v.cancel()
		v.cancel = nil
	}
	v.running = false
	v.addLog("info", "🔌 اتصال قطع شد")
}

func (v *VPNProcess) Status() map[string]interface{} {
	v.mu.Lock()
	defer v.mu.Unlock()
	return map[string]interface{}{
		"running": v.running,
		"pid":     func() int {
			if v.cmd != nil && v.cmd.Process != nil {
				return v.cmd.Process.Pid
			}
			return 0
		}(),
	}
}

func (v *VPNProcess) Logs() []LogLine {
	v.mu.Lock()
	defer v.mu.Unlock()
	out := make([]LogLine, len(v.logs))
	copy(out, v.logs)
	return out
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func classifyLine(line string) string {
	l := strings.ToLower(line)
	switch {
	case strings.Contains(l, "error") || strings.Contains(l, "❌") || strings.Contains(l, "failed"):
		return "error"
	case strings.Contains(l, "warn") || strings.Contains(l, "⚠"):
		return "warn"
	default:
		return "info"
	}
}

func stripANSI(s string) string {
	// Simple ANSI escape remover
	out := strings.Builder{}
	inEsc := false
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEsc = false
			}
			continue
		}
		out.WriteRune(r)
	}
	return out.String()
}

func exeDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func findBinary() string {
	dir := exeDir()
	candidates := []string{
		filepath.Join(dir, "masterdnsvpn-client.exe"),
		filepath.Join(dir, "masterdnsvpn-client-win.exe"),
		filepath.Join(dir, "bin", "masterdnsvpn-client.exe"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// ─── HTTP API handlers ────────────────────────────────────────────────────────

func jsonOK(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}

func jsonErr(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonErr(w, "POST only", 405)
		return
	}
	var req struct {
		BinPath      string `json:"bin_path"`
		ConfigPath   string `json:"config_path"`
		ResolversPath string `json:"resolvers_path"`
		// Inline config fields (optional — we write config file if provided)
		Domain     string `json:"domain"`
		Key        string `json:"key"`
		Method     int    `json:"method"`
		Port       int    `json:"port"`
		Strategy   int    `json:"strategy"`
		Resolvers  string `json:"resolvers"`
		WriteFiles bool   `json:"write_files"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}

	dir := exeDir()
	binPath := req.BinPath
	if binPath == "" {
		binPath = findBinary()
	}
	if binPath == "" {
		jsonErr(w, "باینری masterdnsvpn-client.exe یافت نشد. آن را در کنار این EXE قرار دهید.", 400)
		return
	}

	configPath := req.ConfigPath
	if configPath == "" {
		configPath = filepath.Join(dir, "client_config.toml")
	}
	resolversPath := req.ResolversPath
	if resolversPath == "" {
		resolversPath = filepath.Join(dir, "resolvers.txt")
	}

	// Write config files if user filled in the form
	if req.WriteFiles && req.Domain != "" {
		methodNames := map[int]string{0: "NONE", 1: "XOR", 2: "XChaCha20", 3: "AES-128-GCM", 4: "AES-192-GCM", 5: "AES-256-GCM"}
		_ = methodNames
		port := req.Port
		if port == 0 {
			port = 18000
		}
		strategy := req.Strategy
		if strategy == 0 {
			strategy = 5
		}
		configContent := fmt.Sprintf(`DOMAINS = ["%s"]
DATA_ENCRYPTION_METHOD = %d
ENCRYPTION_KEY = "%s"

PROTOCOL_TYPE = "SOCKS5"
LISTEN_IP = "127.0.0.1"
LISTEN_PORT = %d
SOCKS5_AUTH = false

RESOLVER_BALANCING_STRATEGY = %d
PACKET_DUPLICATION_COUNT = 3
SETUP_PACKET_DUPLICATION_COUNT = 4

LOCAL_DNS_ENABLED = false
LOCAL_DNS_CACHE_MAX_RECORDS = 10000
LOCAL_DNS_CACHE_TTL_SECONDS = 14400.0

MTU_TEST_RETRIES = 2
MTU_TEST_TIMEOUT = 2.0
MTU_TEST_PARALLELISM = 32
MIN_UPLOAD_MTU = 38
MIN_DOWNLOAD_MTU = 200
MAX_UPLOAD_MTU = 150
MAX_DOWNLOAD_MTU = 4000

RX_TX_WORKERS = 4
TUNNEL_PROCESS_WORKERS = 6
ARQ_WINDOW_SIZE = 1000
ARQ_MAX_DATA_RETRIES = 126
LOG_LEVEL = "INFO"
`, req.Domain, req.Method, req.Key, port, strategy)

		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			jsonErr(w, "خطا در نوشتن client_config.toml: "+err.Error(), 500)
			return
		}

		resolversContent := req.Resolvers
		if resolversContent == "" {
			resolversContent = "8.8.8.8\n1.1.1.1\n9.9.9.9"
		}
		_ = os.WriteFile(resolversPath, []byte(resolversContent), 0644)
	}

	if err := vpn.Start(binPath, configPath, resolversPath); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]string{"status": "started"})
}

func handleStop(w http.ResponseWriter, r *http.Request) {
	vpn.Stop()
	jsonOK(w, map[string]string{"status": "stopped"})
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, vpn.Status())
}

func handleLogs(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, vpn.Logs())
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	dir := exeDir()
	bin := findBinary()
	cfgExists := false
	if _, err := os.Stat(filepath.Join(dir, "client_config.toml")); err == nil {
		cfgExists = true
	}
	jsonOK(w, map[string]interface{}{
		"binary_found":   bin != "",
		"binary_path":    bin,
		"config_exists":  cfgExists,
		"exe_dir":        dir,
	})
}

// ─── Window opener (ShellExecute) ─────────────────────────────────────────────

var (
	shell32         = syscall.NewLazyDLL("shell32.dll")
	shellExecute    = shell32.NewProc("ShellExecuteW")
)

func openBrowser(url string) {
	urlPtr, _ := syscall.UTF16PtrFromString(url)
	openPtr, _ := syscall.UTF16PtrFromString("open")
	shellExecute.Call(0,
		uintptr(unsafe.Pointer(openPtr)),
		uintptr(unsafe.Pointer(urlPtr)),
		0, 0,
		uintptr(syscall.SW_SHOWNORMAL),
	)
}

// ─── Find free port ───────────────────────────────────────────────────────────

func freePort() int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 19876
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port
}

// ─── Main ─────────────────────────────────────────────────────────────────────

func main() {
	port := freePort()
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	mux := http.NewServeMux()

	// Serve embedded UI
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "ui/index.html"
		} else {
			path = "ui/" + path
		}
		data, err := uiFiles.ReadFile(path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if strings.HasSuffix(path, ".html") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		} else if strings.HasSuffix(path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		} else if strings.HasSuffix(path, ".css") {
			w.Header().Set("Content-Type", "text/css")
		}
		_, _ = w.Write(data)
	})

	mux.HandleFunc("/api/start", handleStart)
	mux.HandleFunc("/api/stop", handleStop)
	mux.HandleFunc("/api/status", handleStatus)
	mux.HandleFunc("/api/logs", handleLogs)
	mux.HandleFunc("/api/info", handleInfo)

	srv := &http.Server{Addr: addr, Handler: mux}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintln(os.Stderr, "HTTP server error:", err)
		}
	}()

	// Wait a moment for server to be ready then open browser
	time.Sleep(300 * time.Millisecond)
	openBrowser(fmt.Sprintf("http://%s", addr))

	// Keep running until window is closed (no stdin input)
	// A real tray icon would be better here; for now just block
	select {}
}
