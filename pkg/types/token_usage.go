package types

type TokenUsage struct {
	InputTokens     int  `json:"inputTokens,omitempty"`
	OutputTokens    int  `json:"outputTokens,omitempty"`
	CacheReadTokens int  `json:"cacheReadTokens,omitempty"`
	Estimated       bool `json:"estimated,omitempty"`
}

func (u TokenUsage) Total() int {
	return u.InputTokens + u.OutputTokens
}

func (u TokenUsage) HasData() bool {
	return u.InputTokens > 0 || u.OutputTokens > 0 || u.CacheReadTokens > 0
}

func MergeUsage(base TokenUsage, add TokenUsage) TokenUsage {
	base.InputTokens += add.InputTokens
	base.OutputTokens += add.OutputTokens
	base.CacheReadTokens += add.CacheReadTokens
	base.Estimated = base.Estimated || add.Estimated
	return base
}
