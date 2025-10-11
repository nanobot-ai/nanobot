package config

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"sigs.k8s.io/yaml"
)

func TestSchema(t *testing.T) {
	data, err := os.ReadFile("./schema.yaml")
	if err != nil {
		t.Fatalf("Failed to read schema file: %v", err)
	}

	schema := map[string]any{}
	if err := yaml.Unmarshal(data, &schema); err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	c := jsonschema.NewCompiler()
	if err := c.AddResource("schema.json", schema); err != nil {
		t.Fatalf("Failed to add resource to compiler: %v", err)
	}

	s, err := c.Compile("schema.json")
	if err != nil {
		t.Fatalf("Failed to compile schema: %v", err)
	}

	obj := map[string]any{}
	err = json.Unmarshal([]byte(`
{
	"auth": {
		"oauthClientId": "clientid",
		"oauthClientSecret": "clientsecret",
		"oauthAuthorizeUrl": "http://localhost:8080/oauth/authorize",
		"oauthScopes": ["scope1", "scope2"],
		"oauthAuthorizationServerMetadata": {
			"authorizationEndpoint": "http://localhost:8080/oauth/authorize",
			"tokenEndpoint": "http://localhost:8080/oauth/token"
		},
		"encryptionKey": "encryptionkey"
	},
	"mcpServers": {
		"server1": {
			"command": "command1",
			"workdir": "/path/to/workdir",
			"args": ["arg1", "arg2"],
			"url": "http://example.com",
			"env": {
				"env1": "value1",
				"env2": "value2"
			},
			"image": "an image",
			"dockerfile": "Dockerfile content",
			"source": {
				"repo": ".",
				"tag": "latest",
				"commit": "abc123",
				"branch": "main",
				"subPath": "sub/path",
				"reference": "v1.0.0"
			},
			"unsandboxed": true,
			"ports": ["asdf", "fff"],
			"reversePorts": [123,234],
			"headers": {
				"header1": "value1"
			}
		}
	},
	"publish": {
		"name": "test",
		"version": "1.0.0",
		"entrypoint": "entrypoint1",
		"tools": "server1",
		"mcpServers": "server1",
		"instructions": "These are the instructions for the publish.",
		"introduction": "The introduction to the publish.",
		"resources": ["resource1", "resource2"],
		"resourceTemplates": ["resource1", "resource2"],
		"prompts": ["prompt1", "prompt2"]
	},
    "env": {
		"env2": "Short description of env2",
		"env1": {
			"default": "default value",
			"description": "This is the first environment.",
			"options": [ "option1", "option2" ],
			"optional": true,
			"sensitive": true,
			"useBearerToken": true
		}
    },
	"agents": {
		"agent1": {
			"mcpServers": "otherServer",
			"starterMessages": "Hi",
			"icon": "foo",
			"iconDark": "foo",
			"description": "This is the first agent.",
			"model": "a model",
			"tools": "atool",
			"flows": "atool",
			"reasoning": {
				"effort": "low",
				"summary": "detailed"
			},
			"prompts": ["prompt1", "prompt2"],
			"resources": ["prompt1", "prompt2"],
			"agents": "atool",
			"chat": true,
			"instructions": "These are the instructions for the agent.",
			"toolExtensions": {
				"tool1": {
					"extension1": "value1",
					"extension2": true
				}
			},
			"toolChoice": "tool1",
            "temperature": 0.7,
            "topP": 0.2,
			"output": {
				"name": "output1",
				"description": "This is the output schema for agent1.",
				"strict": false,
				"fields": {
					"field1": "description1",
					"field2": "description2",
					"field3": {
						"description": "description3",
						"fields": {
							"fields4": "description4",
							"fields5": "description5"
						}
					}
				}
			},
			"truncation": "auto",
			"maxTokens": 100,
			"aliases": ["alias1", "alias2"],
			"cost": 0.1,
			"speed": 0.3,
			"intelligence": 0.7
		},
		"agent2": {
			"threadName": "a different thread",
			"tools": ["tool1", "tool2"],
			"flows": ["tool1", "tool2"],
			"agents": ["tool1", "tool2"],
			"instructions": {
				"mcpServer": "aserver",
				"prompt": "aprompt",
				"args": {
					"key": "value"
				}
			}
		}
	},
    "flows": {
		"flow2": {
			"input": {
				"description": "This is the input schema for flow1.",
				"schema": {
					"type": "object"
				}
			},
			"steps": [
			]
		},
		"flow1": {
			"before": "target/foo",
			"after": ["target/foo"],
			"description": "This is the first flow.",
			"outputRole": "assistant",
			"input": {
				"name": "a name",
				"description": "This is the input schema for flow1.",
				"fields": {
					"field1": "description1",
					"field2": "description2",
					"field3": {
						"description": "description3",
						"fields": {
							"fields4": "description4",
							"fields5": "description5"
						}
					}
				}
			},
			"steps": [
				{
					"id": "step1",
					"input": {},
					"tool": "tool1",
					"steps": [
						{
							"id": "subStep"
						}
					],
					"else": [
						{
							"id": "subStep"
						}
					]
				},
				{
					"id": "step1",
					"input": "a string",
					"agent": "tool1"
				},
				{
					"id": "elicit1",
					"elicit": "a string"
				},
				{
					"id": "step1",
					"evaluate": "an expression",
					"elicit": {
						"message": "a message",
						"rejectResult": "user declined",
						"cancelResult": {
							"data": "user canceled"
						},
						"input": {
							"fields": {
								"elicit1": "description1"
							}
						}
					}
				},
				{
					"id": "step1",
					"input": "a string",
					"evaluate": {
						"x": "an expression"
					},
					"agent": {
						"name": "tool1",
						"chat": false,
						"newThread": true,
						"temperature": 0.5,
						"topP": 0.5,
						"toolChoice": "tool1",
						"inputAsToolResult": true,
						"output": {
							"description": "output1",
							"fields": {
								"field1": "description1",
								"field1": "description1"
							}
						}
					}
				},
				{
					"id": "step1",
					"input": "a string",
					"flow": "tool1",
					"while": "an expression",
					"return": {
						"key1": {
							"something": 1
						},
						"key2": null,
						"key3": 4
					}
				},
				{
					"id": "step1",
					"input": "a string",
					"flow": "tool1",
					"while": "an expression",
					"set": {
						"key1": {
							"something": 1
						},
						"key2": null,
						"key3": 4
					}
				},
				{
					"id": "step1",
					"parallel": true,
					"forEach": "expr",
					"forEachVar": "item",
					"input": "a string",
					"flow": "tool1"
				},
				{
					"id": "step1",
					"forEach": [
						"item1",
						"item2"
					],
					"forEachVar": "item",
					"input": "a string",
					"flow": "tool1"
				}
			]
		}
	},
	"prompts": {
		"prompt1": {
			"description": "a description",
			"template": "This is the template for prompt1.",
			"input": {
				"field1": "description1",
				"field2": {
					"description": "description2",
					"required": false
				}
			}
		}
	}
}`), &obj)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	err = s.Validate(obj)
	if err != nil {
		t.Fatalf("Failed to validate schema: %v", err)
	}
}
