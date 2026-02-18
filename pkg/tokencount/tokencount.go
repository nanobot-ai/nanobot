package tokencount

import (
	"math"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/modelcaps"
	tiktoken "github.com/pkoukk/tiktoken-go"
	tiktoken_loader "github.com/pkoukk/tiktoken-go-loader"
)

const (
	encodingName = "o200k_base"

	// anthropicMultiplier accounts for the difference between OpenAI's
	// tokenizer and Anthropic's. Tiktoken tends to undercount for Claude
	// models, so we scale up by ~15%.
	anthropicMultiplier = 1.15
)

func init() {
	tiktoken.SetBpeLoader(tiktoken_loader.NewOfflineLoader())
}

var encoding *tiktoken.Tiktoken

func getEncoding() *tiktoken.Tiktoken {
	if encoding == nil {
		enc, err := tiktoken.GetEncoding(encodingName)
		if err != nil {
			// Fall back to nil; CountText will return a chars/4 estimate.
			return nil
		}
		encoding = enc
	}
	return encoding
}

// CountText returns the number of tokens in text using the o200k_base
// encoding. If the encoding is unavailable it falls back to chars/4.
func CountText(text string) int {
	enc := getEncoding()
	if enc == nil {
		return (len([]rune(text)) + 3) / 4
	}
	return len(enc.Encode(text, nil, nil))
}

// CountForModel counts tokens in the concatenated texts and applies a
// provider-specific multiplier based on the model name.
func CountForModel(model string, texts ...string) int {
	total := 0
	for _, t := range texts {
		total += CountText(t)
	}
	return int(math.Ceil(float64(total) * multiplierFor(model)))
}

func multiplierFor(model string) float64 {
	normalized := modelcaps.Normalize(model)
	if strings.HasPrefix(normalized, "claude") {
		return anthropicMultiplier
	}
	return 1.0
}
