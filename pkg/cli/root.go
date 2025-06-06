package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/obot-platform/nanobot/pkg/cmd"
	"github.com/obot-platform/nanobot/pkg/config"
	"github.com/obot-platform/nanobot/pkg/llm"
	"github.com/obot-platform/nanobot/pkg/llm/anthropic"
	"github.com/obot-platform/nanobot/pkg/llm/responses"
	"github.com/obot-platform/nanobot/pkg/log"
	"github.com/obot-platform/nanobot/pkg/runtime"
	"github.com/obot-platform/nanobot/pkg/version"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

func New() *cobra.Command {
	n := &Nanobot{}

	root := cmd.Command(n,
		NewExec(n),
		NewTools(n),
		NewRun(n))
	return root
}

type Nanobot struct {
	Debug            bool              `usage:"Enable debug logging"`
	EmptyEnv         bool              `usage:"Do not load environment variables from the OS"`
	EnvFile          string            `usage:"Path to the environment file (default: ./nanobot.env)" short:"e"`
	OpenAIAPIKey     string            `usage:"OpenAI API key" env:"OPENAI_API_KEY" name:"openai-api-key"`
	OpenAIBaseURL    string            `usage:"OpenAI API URL" env:"OPENAI_BASE_URL" name:"openai-base-url"`
	OpenAIHeaders    map[string]string `usage:"OpenAI API headers" env:"OPENAI_HEADERS" name:"openai-headers"`
	OpenAIAPIVersion string            `usage:"OpenAI API version" env:"OPENAI_API_VERSION" name:"openai-api-version"`
	AnthropicAPIKey  string            `usage:"Anthropic API key" env:"ANTHROPIC_API_KEY" name:"anthropic-api-key"`
	AnthropicBaseURL string            `usage:"Anthropic API URL" env:"ANTHROPIC_BASE_URL" name:"anthropic-base-url"`
	AnthropicHeaders map[string]string `usage:"Anthropic API headers" env:"ANTHROPIC_HEADERS" name:"anthropic-headers"`
}

func (n *Nanobot) Customize(cmd *cobra.Command) {
	cmd.Short = "Nanobot: Build Agents with MCP"
	cmd.Example = `  # Run the example vibe coder Nanobot
  nanobot run vibe-coder

  # Run a nanobot from nanobot.yaml in the local directory
  nanobot run .
`
	cmd.CompletionOptions.HiddenDefaultCmd = true
	cmd.Version = version.Get().String()
}

func (n *Nanobot) PersistentPre(cmd *cobra.Command, _ []string) error {
	if n.Debug {
		log.EnableMessages = true
		log.DebugLog = true
	}

	for _, sub := range cmd.Commands() {
		if sub.Name() == "help" {
			sub.Hidden = true
			sub.Use = " help"
		}
	}
	// Don't need to do anything here, this is just to ensure the env vars get parsed and set always.
	// tbh don't know why this is needed, but it is.
	return nil
}

func display(obj any, format string) bool {
	if format == "json" {
		data, _ := json.MarshalIndent(obj, "", "  ")
		fmt.Println(string(data))
		return true
	} else if format == "yaml" {
		data, _ := yaml.Marshal(obj)
		fmt.Println(string(data))
		return true
	}
	return false
}

func (n *Nanobot) loadEnv() (map[string]string, error) {
	env := map[string]string{}
	if !n.EmptyEnv {
		for _, e := range os.Environ() {
			k, v, _ := strings.Cut(e, "=")
			env[k] = v
		}
	}

	defaultFile := n.EnvFile == ""
	if defaultFile {
		n.EnvFile = "./nanobot.env"
	}

	data, err := os.ReadFile(n.EnvFile)
	if errors.Is(err, fs.ErrNotExist) && defaultFile {
	} else if err != nil {
		return nil, err
	} else {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			k, v, _ := strings.Cut(line, "=")
			env[k] = v
		}
	}

	if _, ok := env["NANOBOT_MCP"]; !ok {
		env["NANOBOT_MCP"] = "true"
	}

	return env, nil
}

func (n *Nanobot) GetRuntime(cfgPath string, opts ...runtime.Options) (*runtime.Runtime, error) {
	cfg, dir, err := config.Load(cfgPath, runtime.CompleteOptions(opts...).Profiles...)
	if err != nil {
		return nil, err
	}
	if dir != "." {
		if err := os.Chdir(dir); err != nil {
			return nil, fmt.Errorf("failed to change directory to %s: %w", dir, err)
		}
	}

	env, err := n.loadEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	return runtime.NewRuntime(env, llm.Config{
		Responses: responses.Config{
			APIKey:  n.OpenAIAPIKey,
			BaseURL: n.OpenAIBaseURL,
			Headers: n.OpenAIHeaders,
			APIVersion: n.OpenAIAPIVersion,
		},
		Anthropic: anthropic.Config{
			APIKey:  n.AnthropicAPIKey,
			BaseURL: n.AnthropicBaseURL,
			Headers: n.AnthropicHeaders,
		},
	}, *cfg, opts...), nil
}

func (n *Nanobot) Run(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}
