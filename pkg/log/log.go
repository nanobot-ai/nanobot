package log

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	EnableMessages    = os.Getenv("NANOBOT_MESSAGES") != "false"
	Base64Replace     = regexp.MustCompile(`(;base64,[a-zA-Z0-9+/=]{0,12})[a-zA-Z0-9+/=]+"`)
	Base64Replacement = []byte(`$1..."`)
)

func Messages(_ context.Context, server string, out bool, data []byte) {
	if !EnableMessages {
		return
	}

	fmtStr := "->(%s) %s\n"
	if !out {
		fmtStr = "<-(%s) %s\n"
	}
	data = Base64Replace.ReplaceAll(data, Base64Replacement)
	_, _ = fmt.Fprintf(os.Stderr, fmtStr, server, strings.ReplaceAll(strings.TrimSpace(string(data)), "\n", " "))
}

func Errorf(_ context.Context, format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
}
