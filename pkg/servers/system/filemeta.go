package system

import (
	"os"
	"time"
)

func buildFileResourceMeta(path string, info os.FileInfo) map[string]any {
	meta := map[string]any{
		"modifiedAt": formatTimestamp(info.ModTime()),
	}

	if createdAt, ok := fileCreatedAt(path, info); ok {
		meta["createdAt"] = formatTimestamp(createdAt)
	}

	return meta
}

func formatTimestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}
