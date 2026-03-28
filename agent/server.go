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
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/usewebtool/webtool/browser"
)

// Server is the daemon that holds the Chrome connection and serves HTTP requests.
type Server struct {
	browser    *browser.Browser
	mu         sync.Mutex
	srv        *http.Server
	logger     *log.Logger
	dir        string        // runtime directory; defaults to runtimeDir(), overridable for tests
	ready      chan struct{} // closed when Connect() completes (success or failure)
	connectErr error         // set before ready is closed; read-only after
}

// NewServer creates a daemon server with the given browser.
func NewServer(b *browser.Browser) *Server {
	return &Server{
		browser: b,
		logger:  log.New(os.Stderr, "", log.LstdFlags),
		dir:     runtimeDir(b.ChromeDataDir),
		ready:   make(chan struct{}),
	}
}

// Start listens on the Unix socket and serves HTTP requests.
// Blocks until shutdown via /stop, SIGTERM, or SIGINT.
func (s *Server) Start() error {
	if s.browser == nil {
		return errors.New("browser is nil")
	}

	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return fmt.Errorf("creating runtime dir: %w", err)
	}

	// Write PID file.
	pid := os.Getpid()
	if err := os.WriteFile(s.pidFile(), []byte(strconv.Itoa(pid)), 0o644); err != nil {
		return fmt.Errorf("writing PID file: %w", err)
	}
	defer s.cleanup()

	// Remove stale socket if it exists.
	os.Remove(s.socketFile())

	ln, err := net.Listen("unix", s.socketFile())
	if err != nil {
		return fmt.Errorf("listening on socket: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("POST /open", s.handleOpen)
	mux.HandleFunc("GET /tabs", s.handleTabs)
	mux.HandleFunc("POST /tab", s.handleSwitch)
	mux.HandleFunc("GET /snapshot", s.handleSnapshot)
	mux.HandleFunc("POST /click", s.handleClick)
	mux.HandleFunc("POST /type", s.handleType)
	mux.HandleFunc("POST /key", s.handleKey)
	mux.HandleFunc("POST /back", s.handleBack)
	mux.HandleFunc("POST /forward", s.handleForward)
	mux.HandleFunc("POST /reload", s.handleReload)

	mux.HandleFunc("POST /eval", s.handleEval)
	mux.HandleFunc("POST /select", s.handleSelect)
	mux.HandleFunc("POST /extract", s.handleExtract)
	mux.HandleFunc("POST /wait", s.handleWait)
	mux.HandleFunc("POST /upload", s.handleUpload)
	mux.HandleFunc("POST /hover", s.handleHover)
	mux.HandleFunc("POST /stop", s.handleStop)

	s.srv = &http.Server{Handler: s.withActionPolicy(mux)}

	// Shut down on SIGTERM/SIGINT.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		<-ctx.Done()
		s.logger.Println("signal received, shutting down")
		s.Shutdown(context.Background())
	}()

	// Connect to Chrome in background. /health blocks on s.ready until this
	// completes, so the client gets the real error if Chrome rejects.
	go func() {
		s.connectErr = s.browser.Connect()
		close(s.ready)
		if s.connectErr != nil {
			s.logger.Printf("chrome connection failed: %v", s.connectErr)
			// Delay shutdown so the client can read the error from /health.
			time.AfterFunc(1*time.Second, func() {
				s.Shutdown(context.Background())
			})
		}
	}()

	s.logger.Printf("daemon started (pid %d), listening on %s", pid, s.socketFile())
	// Blocks until Shutdown() is called (from /stop handler or signal handler).
	err = s.srv.Serve(ln)
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

// alwaysAllowedActions are actions that bypass the action policy.
// These are safe tab management and daemon lifecycle actions.
var alwaysAllowedActions = map[string]bool{
	"health":   true,
	"stop":     true,
	"tabs":     true,
	"tab":      true,
	"snapshot": true,
}

// withActionPolicy rejects requests for actions blocked by the security policy.
// Action names are derived from the URL path (e.g. /click → "click").
// URL paths must match CLI command names and the knownActions set in policy/policy.go.
func (s *Server) withActionPolicy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		action := strings.TrimPrefix(r.URL.Path, "/")
		if alwaysAllowedActions[action] {
			next.ServeHTTP(w, r)
			return
		}
		if !s.browser.Policy().IsActionAllowed(action) {
			writeJSON(w, http.StatusForbidden, Response{
				Error: fmt.Sprintf("action %q is blocked by policy", action),
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	<-s.ready
	if s.connectErr != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Error: s.connectErr.Error()})
		return
	}
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

	if err := s.checkErr(s.browser.Open(r.Context(), req.URL, req.NewTab)); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Error: err.Error()})
		return
	}
	s.logger.Printf("open %s", req.URL)
	writeJSON(w, http.StatusOK, Response{})
}

func (s *Server) handleTabs(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tabs, err := s.browser.Tabs(r.Context())
	if err := s.checkErr(err); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Error: err.Error()})
		return
	}
	s.logger.Println("tabs")
	writeJSON(w, http.StatusOK, TabsResponse{Tabs: tabs})
}

func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	mode := browser.ModeDefault
	switch r.URL.Query().Get("mode") {
	case "interactive":
		mode = browser.ModeInteractive
	case "all":
		mode = browser.ModeAll
	case "", "default":
		// ModeDefault
	default:
		writeJSON(w, http.StatusBadRequest, Response{Error: "mode must be one of: default, interactive, all"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ps, err := s.browser.Snapshot(r.Context(), mode)
	if err := s.checkErr(err); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Error: err.Error()})
		return
	}
	s.logger.Println("snapshot")
	writeJSON(w, http.StatusOK, SnapshotResponse{Snapshot: ps.String()})
}

func (s *Server) handleClick(w http.ResponseWriter, r *http.Request) {
	var req ClickRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Error: fmt.Sprintf("invalid request body: %v", err)})
		return
	}
	if req.Selector == "" {
		writeJSON(w, http.StatusBadRequest, Response{Error: "selector is required"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	el, err := s.browser.Click(r.Context(), req.Selector)
	if err := s.checkErr(err); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Error: err.Error()})
		return
	}
	s.logElement("click", el)
	writeJSON(w, http.StatusOK, Response{})
}

func (s *Server) handleKey(w http.ResponseWriter, r *http.Request) {
	var req KeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Error: fmt.Sprintf("invalid request body: %v", err)})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, Response{Error: "name is required"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.checkErr(s.browser.Key(r.Context(), req.Name)); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Error: err.Error()})
		return
	}
	s.logger.Printf("key %s", req.Name)
	writeJSON(w, http.StatusOK, Response{})
}

func (s *Server) handleType(w http.ResponseWriter, r *http.Request) {
	var req TypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Error: fmt.Sprintf("invalid request body: %v", err)})
		return
	}
	if req.Selector == "" {
		writeJSON(w, http.StatusBadRequest, Response{Error: "selector is required"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	el, err := s.browser.Type(r.Context(), req.Selector, req.Text)
	if err := s.checkErr(err); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Error: err.Error()})
		return
	}
	s.logElement("type", el)
	writeJSON(w, http.StatusOK, Response{})
}

func (s *Server) handleBack(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.checkErr(s.browser.Back(r.Context())); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Error: err.Error()})
		return
	}
	s.logger.Println("back")
	writeJSON(w, http.StatusOK, Response{})
}

func (s *Server) handleForward(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.checkErr(s.browser.Forward(r.Context())); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Error: err.Error()})
		return
	}
	s.logger.Println("forward")
	writeJSON(w, http.StatusOK, Response{})
}

func (s *Server) handleReload(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.checkErr(s.browser.Reload(r.Context())); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Error: err.Error()})
		return
	}
	s.logger.Println("reload")
	writeJSON(w, http.StatusOK, Response{})
}

func (s *Server) handleEval(w http.ResponseWriter, r *http.Request) {
	var req EvalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Error: fmt.Sprintf("invalid request body: %v", err)})
		return
	}
	if req.JS == "" {
		writeJSON(w, http.StatusBadRequest, Response{Error: "js is required"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	result, err := s.browser.Eval(r.Context(), req.JS)
	if err := s.checkErr(err); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Error: err.Error()})
		return
	}
	s.logger.Println("eval")
	writeJSON(w, http.StatusOK, EvalResponse{Result: result})
}

func (s *Server) handleSelect(w http.ResponseWriter, r *http.Request) {
	var req SelectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Error: fmt.Sprintf("invalid request body: %v", err)})
		return
	}
	if req.Selector == "" {
		writeJSON(w, http.StatusBadRequest, Response{Error: "selector is required"})
		return
	}
	if req.Value == "" {
		writeJSON(w, http.StatusBadRequest, Response{Error: "value is required"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	el, err := s.browser.Select(r.Context(), req.Selector, req.Value)
	if err := s.checkErr(err); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Error: err.Error()})
		return
	}
	s.logElement("select", el)
	writeJSON(w, http.StatusOK, Response{})
}

func (s *Server) handleExtract(w http.ResponseWriter, r *http.Request) {
	var req ExtractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Error: fmt.Sprintf("invalid request body: %v", err)})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	content, err := s.browser.Extract(r.Context(), req.Selector, req.AsHTML)
	if err := s.checkErr(err); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Error: err.Error()})
		return
	}
	s.logger.Printf("extract selector=%s", req.Selector)
	writeJSON(w, http.StatusOK, ExtractResponse{Content: content})
}

func (s *Server) handleSwitch(w http.ResponseWriter, r *http.Request) {
	var req SwitchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Error: fmt.Sprintf("invalid request body: %v", err)})
		return
	}
	if req.Index < 1 {
		writeJSON(w, http.StatusBadRequest, Response{Error: "index must be >= 1"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.checkErr(s.browser.Switch(r.Context(), req.Index)); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Error: err.Error()})
		return
	}
	s.logger.Printf("tab %d", req.Index)
	writeJSON(w, http.StatusOK, Response{})
}

func (s *Server) handleWait(w http.ResponseWriter, r *http.Request) {
	var req WaitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Error: fmt.Sprintf("invalid request body: %v", err)})
		return
	}
	if req.Target == "" {
		writeJSON(w, http.StatusBadRequest, Response{Error: "target is required"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.checkErr(s.browser.Wait(r.Context(), req.Target)); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Error: err.Error()})
		return
	}
	s.logger.Printf("wait %s", req.Target)
	writeJSON(w, http.StatusOK, Response{})
}

func (s *Server) handleHover(w http.ResponseWriter, r *http.Request) {
	var req HoverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Error: fmt.Sprintf("invalid request body: %v", err)})
		return
	}
	if req.Selector == "" {
		writeJSON(w, http.StatusBadRequest, Response{Error: "selector is required"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	el, err := s.browser.Hover(r.Context(), req.Selector)
	if err := s.checkErr(err); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Error: err.Error()})
		return
	}
	s.logElement("hover", el)
	writeJSON(w, http.StatusOK, Response{})
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	var req UploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Error: fmt.Sprintf("invalid request body: %v", err)})
		return
	}
	if req.Selector == "" {
		writeJSON(w, http.StatusBadRequest, Response{Error: "selector is required"})
		return
	}
	if len(req.Files) == 0 {
		writeJSON(w, http.StatusBadRequest, Response{Error: "files is required"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	el, err := s.browser.Upload(r.Context(), req.Selector, req.Files)
	if err := s.checkErr(err); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Error: err.Error()})
		return
	}
	s.logElement("upload", el)
	writeJSON(w, http.StatusOK, Response{})
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	s.logger.Println("stop")
	writeJSON(w, http.StatusOK, Response{})
	go s.Shutdown(context.Background())
}

// logElement logs the accessible role and name of the element that was acted on.
func (s *Server) logElement(action string, el *browser.Element) {
	if el.Role != "" {
		s.logger.Printf("%s %s %q", action, el.Role, el.Name)
	}
}

// checkErr checks for async policy errors from the browser's active tab.
// If a policy error is present, it overrides the original error.
// Must be called inside the mutex.
func (s *Server) checkErr(err error) error {
	if errTab := s.browser.Err(); errTab != nil {
		return errTab
	}
	return browser.FriendlyError(err)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
