package config

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"github.com/obot-platform/nanobot/pkg/log"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"sigs.k8s.io/yaml"
)

var (
	schemaOnce sync.Once
	schema     *jsonschema.Schema
	//go:embed schema.yaml
	schemaByte []byte
)

func getSchema() *jsonschema.Schema {
	schemaOnce.Do(func() {
		var err error
		schema, err = initSchema()
		if err != nil {
			log.Fatalf(context.Background(), "error initializing schema: %v", err)
		}
	})
	return schema
}

func initSchema() (*jsonschema.Schema, error) {
	schemaObj := map[string]any{}
	if err := yaml.Unmarshal(schemaByte, &schemaObj); err != nil {
		return nil, err
	}

	c := jsonschema.NewCompiler()
	if err := c.AddResource("schema.json", schemaObj); err != nil {
		return nil, fmt.Errorf("error adding schema resource: %w", err)
	}

	s, err := c.Compile("schema.json")
	if err != nil {
		return nil, fmt.Errorf("error compiling schema: %w", err)
	}

	return s, nil
}
