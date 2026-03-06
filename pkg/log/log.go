package log

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"slices"
	"strings"
)

var (
	debugs            = strings.Split(os.Getenv("NANOBOT_DEBUG"), ",")
	EnableMessages    = slices.Contains(debugs, "messages")
	EnableProgress    = slices.Contains(debugs, "progress")
	EnableUI          = slices.Contains(debugs, "ui")
	Base64Replace     = regexp.MustCompile(`((;base64,|")[a-zA-Z0-9+/=]{60})[a-zA-Z0-9+/=]+"`)
	Base64Replacement = []byte(`$1..."`)
)

func init() {
	ConfigureSlog(false, false)
}

func ConfigureSlog(debug, trace bool) {
	level := slog.LevelInfo
	if slices.Contains(debugs, "trace") || slices.Contains(debugs, "log") || debug || trace {
		level = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})))
}

func Messages(_ context.Context, server string, out bool, data []byte) {
	if !EnableUI && server == "nanobot.ui" {
		return
	}

	if EnableProgress && bytes.Contains(data, []byte(`"notifications/progress"`)) {
	} else if EnableMessages && !bytes.Contains(data, []byte(`"notifications/progress"`)) {
	} else if slices.Contains(debugs, server) {
	} else {
		return
	}

	prefixFmt := "->(%s)"
	if !out {
		prefixFmt = "<-(%s)"
	}

	data = Base64Replace.ReplaceAll(data, Base64Replacement)
	slog.Info("mcp message", "prefix", fmt.Sprintf(prefixFmt, server), "payload", strings.ReplaceAll(strings.TrimSpace(string(data)), "\n", " "))
}

func StderrMessages(_ context.Context, server, line string) {
	slog.Info("mcp stderr", "server", server, "stream", "stderr", "line", strings.TrimRight(line, "\n"))
}
