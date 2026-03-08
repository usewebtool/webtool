package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/machinae/webtool/browser"
)

// Server is the daemon that holds the Chrome connection and serves HTTP requests.
type Server struct {
	browser *browser.Browser
	mu      sync.Mutex
	srv     *http.Server
	logger  *log.Logger
	dir     string // runtime directory; defaults to runtimeDir(), overridable for tests
}

// NewServer creates a daemon server with the given browser.
func NewServer(b *browser.Browser) *Server {
	return &Server{
		browser: b,
		logger:  log.New(os.Stderr, "", log.LstdFlags),
		dir:     runtimeDir(b.ChromeDataDir),
	}
}

// Start listens on the Unix socket and serves HTTP requests.
// Blocks until shutdown via /stop, SIGTERM, or SIGINT.
func (s *Server) Start() error {
	if s.browser == nil {
		return errors.New("browser is nil")
	}

	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("creating runtime dir: %w", err)
	}

	// Write PID file.
	pid := os.Getpid()
	if err := os.WriteFile(s.pidFile(), []byte(strconv.Itoa(pid)), 0o644); err != nil {
		return fmt.Errorf("writing PID file: %w", err)
	}

	// Remove stale socket if it exists.
	os.Remove(s.socketFile())

	ln, err := net.Listen("unix", s.socketFile())
	if err != nil {
		os.Remove(s.pidFile())
		return fmt.Errorf("listening on socket: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("POST /open", s.handleOpen)
	mux.HandleFunc("GET /tabs", s.handleTabs)
	mux.HandleFunc("POST /stop", s.handleStop)

	s.srv = &http.Server{Handler: mux}

	// Shut down on SIGTERM/SIGINT.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		<-ctx.Done()
		s.logger.Println("signal received, shutting down")
		s.Shutdown(context.Background())
	}()

	s.logger.Printf("daemon started (pid %d), listening on %s", pid, s.socketFile())
	// Blocks until Shutdown() is called (from /stop handler or signal handler).
	err = s.srv.Serve(ln)

	s.cleanup()

	// Serve returns ErrServerClosed on graceful shutdown — not a real error.
	if err == http.ErrServerClosed {
		err = nil
	}
	return err
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.srv != nil {
		return s.srv.Shutdown(ctx)
	}
	return nil
}

func (s *Server) cleanup() {
	s.browser.Close()
	os.Remove(s.socketFile())
	os.Remove(s.pidFile())
	s.logger.Println("daemon stopped")
}

func (s *Server) socketFile() string {
	return s.dir + "/agent.sock"
}

func (s *Server) pidFile() string {
	return s.dir + "/agent.pid"
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, Response{})
}

func (s *Server) handleOpen(w http.ResponseWriter, r *http.Request) {
	var req OpenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Error: fmt.Sprintf("invalid request body: %v", err)})
		return
	}
	if req.URL == "" {
		writeJSON(w, http.StatusBadRequest, Response{Error: "url is required"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.browser.Open(r.Context(), req.URL); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, Response{})
}

func (s *Server) handleTabs(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tabs, err := s.browser.Tabs(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, TabsResponse{Tabs: tabs})
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, Response{})
	go s.Shutdown(context.Background())
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
