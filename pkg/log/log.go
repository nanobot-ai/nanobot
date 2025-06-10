package log

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/obot-platform/nanobot/pkg/printer"
)

var (
	debugs            = strings.Split(os.Getenv("NANOBOT_DEBUG"), ",")
	EnableMessages    = slices.Contains(debugs, "messages")
	EnableProgress    = slices.Contains(debugs, "progress")
	DebugLog          = slices.Contains(debugs, "log")
	Base64Replace     = regexp.MustCompile(`((;base64,|")[a-zA-Z0-9+/=]{60})[a-zA-Z0-9+/=]+"`)
	Base64Replacement = []byte(`$1..."`)
)

func Messages(_ context.Context, server string, out bool, data []byte) {
	if EnableProgress && bytes.Contains(data, []byte(`"notifications/progress"`)) {
	} else if EnableMessages && !bytes.Contains(data, []byte(`"notifications/progress"`)) {
	} else {
		return
	}

	prefixFmt := "->(%s)"
	if !out {
		prefixFmt = "<-(%s)"
	}
	data = Base64Replace.ReplaceAll(data, Base64Replacement)
	printer.Prefix(fmt.Sprintf(prefixFmt, server), strings.ReplaceAll(strings.TrimSpace(string(data)), "\n", " ")+"\n")
}

func StderrMessages(_ context.Context, server, line string) {
	printer.Prefix(fmt.Sprintf("<-(%s:stderr)", server), line+"\n")
}

func Errorf(_ context.Context, format string, args ...any) {
	printer.Prefix("ERROR:", fmt.Sprintf(format+"\n", args...))
}

func Infof(_ context.Context, format string, args ...any) {
	printer.Prefix("INFO:", fmt.Sprintf(format+"\n", args...))
}

func Fatalf(_ context.Context, format string, args ...any) {
	printer.Prefix("FATAL:", fmt.Sprintf(format+"\n", args...))
	os.Exit(1)
}

func Debugf(_ context.Context, format string, args ...any) {
	if !DebugLog {
		return
	}
	printer.Prefix("DEBUG:", fmt.Sprintf(format+"\n", args...))
}
