package types

import (
	"strings"
	"unicode/utf8"
)

const (
	MessageMimeType     = "application/vnd.nanobot.chat.message+json"
	HistoryMimeType     = "application/vnd.nanobot.chat.history+json"
	ToolResultMimeType  = "application/vnd.nanobot.tool.result+json"
	ErrorMimeType       = "application/vnd.nanobot.error+json"
	AgentMimeType       = "application/vnd.nanobot.agent+json"
	SessionMimeType     = "application/vnd.nanobot.session+json"
	ElicitationMimeType = "application/vnd.nanobot.elicitation+json"
	MetaNanobot         = "ai.nanobot"

	MessageURI     = "chat://message/%s"
	HistoryURI     = "chat://history"
	ProgressURI    = "chat://progress"
	ElicitationURI = "chat://elicitation"
)

var (
	ImageMimeTypes = map[string]struct{}{
		"image/png":  {},
		"image/jpeg": {},
		"image/webp": {},
	}
	TextMimeTypes = map[string]struct{}{
		"text/plain":             {},
		"text/markdown":          {},
		"text/html":              {},
		"text/csv":               {},
		"application/json":       {},
		"application/xml":        {},
		"application/javascript": {},
		"image/svg+xml":          {},
	}
	PDFMimeTypes = map[string]struct{}{
		"application/pdf": {},
	}
)

// ResourceContentUseBlob reports whether file bytes should be sent in MCP ResourceContent.blob
// (base64) rather than ResourceContent.text (UTF-8 string). Binary formats such as Office
// documents must use blob; interpreting them as Go strings corrupts the payload.
func ResourceContentUseBlob(mimeType string, content []byte) bool {
	if strings.HasPrefix(mimeType, "text/") {
		return false
	}
	if _, ok := TextMimeTypes[mimeType]; ok {
		return false
	}
	if _, ok := ImageMimeTypes[mimeType]; ok {
		return true
	}
	if _, ok := PDFMimeTypes[mimeType]; ok {
		return true
	}
	// OOXML (docx, xlsx, pptx, templates, etc.) is always binary (ZIP package).
	if strings.HasPrefix(mimeType, "application/vnd.openxmlformats") {
		return true
	}
	if mimeType == "application/zip" {
		return true
	}
	if strings.HasPrefix(mimeType, "audio/") || strings.HasPrefix(mimeType, "video/") || strings.HasPrefix(mimeType, "font/") {
		return true
	}
	// Other image/* (e.g. gif) — SVG is listed in TextMimeTypes above.
	if strings.HasPrefix(mimeType, "image/") {
		return true
	}
	if mimeType == "application/octet-stream" || mimeType == "" {
		return !utf8.ValidString(string(content))
	}
	if strings.HasPrefix(mimeType, "application/") {
		return !utf8.ValidString(string(content))
	}
	return true
}
