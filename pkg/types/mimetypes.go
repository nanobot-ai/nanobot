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

	AsyncMetaKey = "ai.nanobot.async"
)

var (
	ImageMimeTypes = map[string]struct{}{
		"image/png":  {},
		"image/jpeg": {},
		"image/webp": {},
	}
	TextMimeTypes = map[string]struct{}{
		"text/plain":       {},
		"text/markdown":    {},
		"text/html":        {},
		"text/csv":         {},
		"application/json": {},
	}
	PDFMimeTypes = map[string]struct{}{
		"application/pdf": {},
	}
)

func Meta(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	return map[string]any{MetaNanobot: m}
}
