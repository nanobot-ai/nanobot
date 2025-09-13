package types

const (
	MessageMimeType    = "application/vnd.nanobot.chat.message+json"
	HistoryMimeType    = "application/vnd.nanobot.chat.history+json"
	ToolResultMimeType = "application/vnd.nanobot.tool.result+json"
	ErrorMimeType      = "application/vnd.nanobot.error+json"

	MessageURI  = "chat://message/%s"
	HistoryURI  = "chat://history"
	ProgressURI = "chat://progress"

	AsyncMetaKey = "ai.nanobot.async"
)
