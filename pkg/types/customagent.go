package types

type CustomAgentIcons struct {
	Light string `json:"light"`
	Dark  string `json:"dark"`
}

type CustomAgent struct {
	CustomAgentMeta
	RemoteURL           string                 `json:"remoteUrl,omitempty"`
	Icons               *CustomAgentIcons      `json:"icons"`
	BaseAgent           string                 `json:"baseAgent,omitempty"`
	Instructions        string                 `json:"instructions"`
	IntroductionMessage string                 `json:"introductionMessage"`
	StarterMessages     []string               `json:"starterMessages"`
	KnowledgeResources  []string               `json:"knowledgeResources,omitempty"`
	MCPServers          []CustomAgentMCPServer `json:"mcpServers,omitempty"`
}

type CustomAgentMCPServer struct {
	URL            string   `json:"url"`
	EnabledTools   []string `json:"enabledTools"`
	EnabledPrompts []string `json:"enabledPrompts,omitempty"`
}

type CustomAgentMeta struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"isPublic,omitempty"` // Indicates if the agent can be shared with others
}
