// ==============================================================================
// MasterDnsVPN — Innovation: Metrics HTTP Server
// Exposes /metrics (Prometheus text format) and /api/status (JSON).
// Zero external dependencies — uses only the standard library.
// ==============================================================================
package metrics

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

var (
	BytesUpTotal    atomic.Int64
	BytesDownTotal  atomic.Int64
	PacketsDropped  atomic.Int64
	ActiveStreams    atomic.Int32
	ActiveResolvers atomic.Int32
	AvgLatencyMs    atomic.Int64
	SessionResets   atomic.Int64
	startTime       = time.Now()
)

type Server struct {
	addr string
	srv  *http.Server
}

func New(addr string) *Server {
	mux := http.NewServeMux()
	s := &Server{addr: addr}
	mux.HandleFunc("/metrics", s.handlePrometheus)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ok")
	})
	s.srv = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	return s
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("metrics listen %s: %w", s.addr, err)
	}
	go func() { _ = s.srv.Serve(ln) }()
	return nil
}

func (s *Server) Stop() { _ = s.srv.Close() }

func (s *Server) handlePrometheus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	upSec := int64(time.Since(startTime).Seconds())
	fmt.Fprintf(w, "masterdnsvpn_uptime_seconds %d\n", upSec)
	fmt.Fprintf(w, "masterdnsvpn_bytes_up_total %d\n", BytesUpTotal.Load())
	fmt.Fprintf(w, "masterdnsvpn_bytes_down_total %d\n", BytesDownTotal.Load())
	fmt.Fprintf(w, "masterdnsvpn_packets_dropped_total %d\n", PacketsDropped.Load())
	fmt.Fprintf(w, "masterdnsvpn_active_streams %d\n", ActiveStreams.Load())
	fmt.Fprintf(w, "masterdnsvpn_active_resolvers %d\n", ActiveResolvers.Load())
	fmt.Fprintf(w, "masterdnsvpn_avg_latency_ms %d\n", AvgLatencyMs.Load())
	fmt.Fprintf(w, "masterdnsvpn_session_resets_total %d\n", SessionResets.Load())
}

type statusResponse struct {
	UptimeSec       int64  `json:"uptime_sec"`
	BytesUp         int64  `json:"bytes_up_total"`
	BytesDown       int64  `json:"bytes_down_total"`
	PacketsDropped  int64  `json:"packets_dropped_total"`
	ActiveStreams    int32  `json:"active_streams"`
	ActiveResolvers int32  `json:"active_resolvers"`
	AvgLatencyMs    int64  `json:"avg_latency_ms"`
	SessionResets   int64  `json:"session_resets_total"`
	StartTime       string `json:"start_time"`
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := statusResponse{
		UptimeSec:       int64(time.Since(startTime).Seconds()),
		BytesUp:         BytesUpTotal.Load(),
		BytesDown:       BytesDownTotal.Load(),
		PacketsDropped:  PacketsDropped.Load(),
		ActiveStreams:    ActiveStreams.Load(),
		ActiveResolvers: ActiveResolvers.Load(),
		AvgLatencyMs:    AvgLatencyMs.Load(),
		SessionResets:   SessionResets.Load(),
		StartTime:       startTime.UTC().Format(time.RFC3339),
	}
	_ = json.NewEncoder(w).Encode(resp)
}
