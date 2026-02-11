package modelcaps

import "strings"

type Limits struct {
	Context       int
	OutputReserve int
}

func ContextWindow(modelID string) int {
	return lookup(Normalize(modelID)).Context
}

func ReservedOutput(modelID string) int {
	return lookup(Normalize(modelID)).OutputReserve
}

func InputCap(modelID string) int {
	model := Normalize(modelID)
	limits := lookup(model)
	if limits.Context == 0 {
		return 0
	}

	usable := limits.Context - limits.OutputReserve
	if usable < 0 {
		usable = 0
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

func Normalize(modelID string) string {
	model := strings.ToLower(strings.TrimSpace(modelID))
	if idx := strings.LastIndex(model, "/"); idx >= 0 {
		model = model[idx+1:]
	}
	if idx := strings.LastIndex(model, ":"); idx >= 0 {
		model = model[idx+1:]
	}
	return model
}

func anthropicLimits(_ string) Limits {
	return Limits{Context: 200_000, OutputReserve: 5_000}
}

func gpt5Limits(_ string) Limits {
	return Limits{Context: 272_000, OutputReserve: 5_000}
}
