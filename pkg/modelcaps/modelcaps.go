package modelcaps

import "strings"

type Limits struct {
	Context       int
	OutputReserve int
	Input         int
}

func ContextWindow(modelID string) int {
	return lookup(normalize(modelID)).Context
}

func ReservedOutput(modelID string) int {
	return lookup(normalize(modelID)).OutputReserve
}

func InputCap(modelID string) int {
	model := normalize(modelID)
	limits := lookup(model)
	if limits.Context == 0 {
		return 0
	}

	usable := limits.Context - limits.OutputReserve
	if usable < 0 {
		usable = 0
	}

	if limits.Input > 0 && limits.Input < usable {
		usable = limits.Input
	}

	return usable
}

func lookup(model string) Limits {
	switch {
	case strings.HasPrefix(model, "claude"):
		return anthropicLimits(model)
	case strings.HasPrefix(model, "gpt-5"):
		return gpt5Limits(model)
	default:
		return Limits{}
	}
}

func normalize(modelID string) string {
	model := strings.ToLower(strings.TrimSpace(modelID))
	if idx := strings.LastIndex(model, "/"); idx >= 0 {
		model = model[idx+1:]
	}
	if idx := strings.LastIndex(model, ":"); idx >= 0 {
		model = model[idx+1:]
	}
	return model
}

func anthropicLimits(model string) Limits {
	limits := Limits{Context: 200_000, OutputReserve: 8_000}

	switch {
	case strings.Contains(model, "opus-4"):
		limits.OutputReserve = 32_000
	case strings.Contains(model, "sonnet-4") || strings.Contains(model, "haiku-4"):
		limits.OutputReserve = 64_000
	case strings.Contains(model, "opus"):
		limits.OutputReserve = 4_096
	}

	return limits
}

func gpt5Limits(model string) Limits {
	switch {
	case strings.Contains(model, "pro"):
		return Limits{
			Context:       400_000,
			OutputReserve: 272_000,
			Input:         272_000,
		}
	case strings.Contains(model, "chat"):
		return Limits{
			Context:       128_000,
			OutputReserve: 16_384,
		}
	default:
		return Limits{
			Context:       272_000,
			OutputReserve: 128_000,
		}
	}
}
