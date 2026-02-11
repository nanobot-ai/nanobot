package contextguard

import (
	"encoding/json"
	"math"
	"unicode/utf8"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/modelcaps"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

type Status string

const (
	StatusOK              Status = "ok"
	StatusNeedsCompaction Status = "needs_compaction"
	StatusOverLimit       Status = "over_limit"
)

type Limits struct {
	Context  int
	Reserved int
	Usable   int
}

type Totals struct {
	InputTokens int
	ToolTokens  int
	Estimated   bool
}

type Result struct {
	Status Status
	Totals Totals
	Limits Limits
}

type Config struct {
	WarnThreshold float64
}

type Service struct {
	config Config
}

type State struct {
	Model        string
	SystemPrompt string
	Tools        []types.ToolUseDefinition
	Messages     []types.Message
}

const defaultWarnThreshold = 0.9

func NewService(cfg Config) Service {
	threshold := cfg.WarnThreshold
	if threshold == 0 {
		threshold = defaultWarnThreshold
	}
	return Service{config: Config{WarnThreshold: threshold}}
}

func (s Service) Evaluate(state State) Result {
	limits := Limits{
		Context:  modelcaps.ContextWindow(state.Model),
		Reserved: modelcaps.ReservedOutput(state.Model),
		Usable:   modelcaps.InputCap(state.Model),
	}

	totals := estimateTotals(state)

	status := StatusOK
	if limits.Usable > 0 {
		ratio := float64(totals.InputTokens) / float64(limits.Usable)
		switch {
		case totals.InputTokens >= limits.Usable:
			status = StatusOverLimit
		case ratio >= s.config.WarnThreshold:
			status = StatusNeedsCompaction
		default:
			status = StatusOK
		}
	}

	return Result{
		Status: status,
		Totals: totals,
		Limits: limits,
	}
}

func estimateTotals(state State) Totals {
	totalChars := utf8.RuneCountInString(state.SystemPrompt)
	toolChars := 0

	for _, tool := range state.Tools {
		totalChars += toolDefinitionChars(tool)
	}

	for _, msg := range state.Messages {
		msgChars, msgToolChars := messageChars(msg)
		totalChars += msgChars
		toolChars += msgToolChars
	}

	return Totals{
		InputTokens: toTokens(totalChars),
		ToolTokens:  toTokens(toolChars),
		Estimated:   true,
	}
}

func toolDefinitionChars(tool types.ToolUseDefinition) int {
	chars := utf8.RuneCountInString(tool.Name)
	chars += utf8.RuneCountInString(tool.Description)
	if len(tool.Parameters) > 0 {
		chars += utf8.RuneCountInString(string(tool.Parameters))
	}
	return chars
}

func messageChars(msg types.Message) (total, tool int) {
	for _, item := range msg.Items {
		if item.Content != nil {
			total += contentChars(*item.Content)
		}
		if item.ToolCall != nil {
			total += utf8.RuneCountInString(item.ToolCall.Arguments)
		}
		if item.ToolCallResult != nil {
			chars := callResultChars(item.ToolCallResult.Output)
			total += chars
			tool += chars
		}
		if item.Reasoning != nil {
			for _, summary := range item.Reasoning.Summary {
				total += utf8.RuneCountInString(summary.Text)
			}
		}
	}
	return total, tool
}

func callResultChars(result types.CallResult) int {
	chars := 0
	for _, content := range result.Content {
		chars += contentChars(content)
	}
	return chars
}

func contentChars(content mcp.Content) int {
	chars := utf8.RuneCountInString(content.Text)

	if content.Resource != nil {
		chars += utf8.RuneCountInString(content.Resource.Text)
		if content.Resource.Blob != "" {
			chars += approxBase64Chars(content.Resource.Blob)
		}
	}

	if content.Data != "" {
		chars += approxBase64Chars(content.Data)
	}

	for _, child := range content.Content {
		chars += contentChars(child)
	}

	if len(content.StructuredContent) > 0 {
		if data, err := json.Marshal(content.StructuredContent); err == nil {
			chars += utf8.RuneCountInString(string(data))
		}
	}

	return chars
}

func approxBase64Chars(data string) int {
	return int(math.Round(float64(len(data)) * 0.75))
}

func toTokens(chars int) int {
	if chars <= 0 {
		return 0
	}
	return (chars + 3) / 4
}
