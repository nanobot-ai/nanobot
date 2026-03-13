package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultBrowserProxyTarget = "http://127.0.0.1:6080"
	defaultBrowserDisplay     = ":99"
	defaultBrowserMaxRes      = "4096x4096"
	minBrowserWidth           = 640
	minBrowserHeight          = 480
	browserResizeRounding     = 8
)

type browserResizeRequest struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type browserStatusResponse struct {
	Available   bool `json:"available"`
	WindowCount int  `json:"windowCount"`
}

type browserProxyHandler struct {
	proxy     http.Handler
	display   string
	maxWidth  int
	maxHeight int

	mu          sync.Mutex
	currentSize string
}

func newBrowserProxy() http.Handler {
	target, err := url.Parse(envOrDefault("BROWSER_WEBSOCKET_TARGET", defaultBrowserProxyTarget))
	if err != nil {
		panic("invalid BROWSER_WEBSOCKET_TARGET: " + err.Error())
	}

	maxWidth, maxHeight, err := parseResolution(envOrDefault("BROWSER_VIEW_MAX_RESOLUTION", defaultBrowserMaxRes))
	if err != nil {
		panic("invalid BROWSER_VIEW_MAX_RESOLUTION: " + err.Error())
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = trimBrowserPrefix(req.URL.Path)
		req.URL.RawPath = trimBrowserPrefix(req.URL.RawPath)
	}

	return &browserProxyHandler{
		proxy:     proxy,
		display:   envOrDefault("DISPLAY", defaultBrowserDisplay),
		maxWidth:  maxWidth,
		maxHeight: maxHeight,
	}
}

func (h *browserProxyHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	switch trimBrowserPrefix(req.URL.Path) {
	case "/healthz":
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("ok"))
	case "/status":
		h.handleStatusStream(rw, req)
	case "/resize":
		h.handleResize(rw, req)
	default:
		h.proxy.ServeHTTP(rw, req)
	}
}

func (h *browserProxyHandler) handleStatusStream(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		rw.Header().Set("Allow", http.MethodGet)
		http.Error(rw, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	flusher, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("X-Accel-Buffering", "no")

	writeStatus := func(status browserStatusResponse) error {
		payload, err := json.Marshal(status)
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintf(rw, "event: status\ndata: %s\n\n", payload); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	lastStatus := h.currentStatus()
	if err := writeStatus(lastStatus); err != nil {
		return
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-req.Context().Done():
			return
		case <-ticker.C:
			status := h.currentStatus()
			if status == lastStatus {
				if _, err := rw.Write([]byte(": keepalive\n\n")); err != nil {
					return
				}
				flusher.Flush()
				continue
			}

			if err := writeStatus(status); err != nil {
				return
			}
			lastStatus = status
		}
	}
}

func (h *browserProxyHandler) handleResize(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		rw.Header().Set("Allow", http.MethodPost)
		http.Error(rw, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload browserResizeRequest
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		http.Error(rw, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	width, height := h.normalizeSize(payload.Width, payload.Height)
	if width == 0 || height == 0 {
		http.Error(rw, "width and height must be positive", http.StatusBadRequest)
		return
	}

	if err := h.resize(width, height); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(rw).Encode(map[string]int{
		"width":  width,
		"height": height,
	})
}

func (h *browserProxyHandler) normalizeSize(width, height int) (int, int) {
	if width <= 0 || height <= 0 {
		return 0, 0
	}

	width = maxInt(width, minBrowserWidth)
	height = maxInt(height, minBrowserHeight)
	width = minInt(width, h.maxWidth)
	height = minInt(height, h.maxHeight)

	width = roundUp(width, browserResizeRounding)
	height = roundUp(height, browserResizeRounding)

	return width, height
}

func (h *browserProxyHandler) resize(width, height int) error {
	sizeKey := strconv.Itoa(width) + "x" + strconv.Itoa(height)

	h.mu.Lock()
	defer h.mu.Unlock()

	if sizeKey != h.currentSize {
		cmd := exec.Command("xrandr", "--display", h.display, "--fb", sizeKey)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return errors.New("xrandr resize failed: " + strings.TrimSpace(string(output)))
		}

		h.currentSize = sizeKey
	}

	if err := h.maximizeWindows(); err != nil {
		return err
	}

	return nil
}

func (h *browserProxyHandler) currentStatus() browserStatusResponse {
	windowCount, _ := h.browserWindowCount()
	return browserStatusResponse{
		Available:   windowCount > 0,
		WindowCount: windowCount,
	}
}

func (h *browserProxyHandler) maximizeWindows() error {
	listCmd := exec.Command("wmctrl", "-l")
	listOutput, err := listCmd.Output()
	if err != nil {
		return errors.New("wmctrl list failed: " + strings.TrimSpace(string(listOutput)))
	}

	for _, line := range strings.Split(string(listOutput), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		windowID := fields[0]
		if output, err := exec.Command("wmctrl", "-ir", windowID, "-b", "add,maximized_vert,maximized_horz").CombinedOutput(); err != nil {
			return errors.New("wmctrl maximize failed: " + strings.TrimSpace(string(output)))
		}
		if output, err := exec.Command("wmctrl", "-ir", windowID, "-e", "0,0,0,-1,-1").CombinedOutput(); err != nil {
			return errors.New("wmctrl geometry failed: " + strings.TrimSpace(string(output)))
		}
	}

	return nil
}

func (h *browserProxyHandler) browserWindowCount() (int, error) {
	output, err := exec.Command("wmctrl", "-lx").CombinedOutput()
	if err != nil {
		return 0, err
	}

	return countBrowserWindows(string(output)), nil
}

func countBrowserWindows(output string) int {
	count := 0
	for _, line := range strings.Split(output, "\n") {
		lower := strings.ToLower(line)
		if lower == "" {
			continue
		}

		if strings.Contains(lower, "chromium") || strings.Contains(lower, "chrome") {
			count++
		}
	}

	return count
}

func parseResolution(value string) (int, int, error) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(value)), "x")
	if len(parts) != 2 {
		return 0, 0, errors.New("resolution must be WIDTHxHEIGHT")
	}

	width, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	height, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	if width <= 0 || height <= 0 {
		return 0, 0, errors.New("resolution values must be positive")
	}

	return width, height, nil
}

func trimBrowserPrefix(path string) string {
	if path == "" {
		return path
	}
	trimmed := strings.TrimPrefix(path, "/browser")
	if trimmed == "" {
		return "/"
	}
	return trimmed
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func roundUp(value, step int) int {
	if step <= 1 {
		return value
	}
	remainder := value % step
	if remainder == 0 {
		return value
	}
	return value + (step - remainder)
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
