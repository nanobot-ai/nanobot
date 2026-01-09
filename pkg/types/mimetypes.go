package types

const (
	MessageMimeType    = "application/vnd.nanobot.chat.message+json"
	HistoryMimeType    = "application/vnd.nanobot.chat.history+json"
	ToolResultMimeType = "application/vnd.nanobot.tool.result+json"
	ErrorMimeType      = "application/vnd.nanobot.error+json"
	AgentMimeType      = "application/vnd.nanobot.agent+json"
	WorkspaceMimeType  = "application/vnd.nanobot.workspace+json"
	SessionMimeType    = "application/vnd.nanobot.session+json"
	MetaNanobot        = "ai.nanobot"

	MessageURI  = "chat://message/%s"
	HistoryURI  = "chat://history"
	ProgressURI = "chat://progress"
)

var (
	ImageMimeTypes = map[string]struct{}{
		"image/png":  {},
		"image/jpeg": {},
		"image/webp": {},
		"image/gif":  {},
	}
	TextMimeTypes = map[string]struct{}{
		"text/plain":               {},
		"text/markdown":            {},
		"text/html":                {},
		"text/csv":                 {},
		"text/yaml":                {},
		"text/css":                 {},
		"text/javascript":          {},
		"text/typescript":          {},
		"text/x-python":            {},
		"text/x-go":                {},
		"text/x-rust":              {},
		"text/x-java":              {},
		"text/x-c":                 {},
		"text/x-c++":               {},
		"text/x-shellscript":       {},
		"application/json":         {},
		"application/xml":          {},
		"application/javascript":   {},
		"application/octet-stream": {}, // Treat unknown as text to avoid image_url errors
	}
	PDFMimeTypes = map[string]struct{}{
		"application/pdf": {},
	}
)
