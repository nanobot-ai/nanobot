package types

type Dialect string

const (
	DialectAnthropicMessages      Dialect = "AnthropicMessages"
	DialectOpenAIResponses        Dialect = "OpenAIResponses"
	DialectOpenResponses          Dialect = "OpenResponses"
	DialectOpenAIChatCompletions  Dialect = "OpenAIChatCompletions"
	DialectDefault                        = DialectOpenAIResponses
)
