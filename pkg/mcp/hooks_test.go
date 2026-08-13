package mcp

import (
	"encoding/json"
	"testing"

	filtercontract "github.com/obot-platform/nanobot/pkg/filter"
)

func TestHookTargetContractVersionSerialization(t *testing.T) {
	tests := []struct {
		name   string
		target HookTarget
		want   string
	}{
		{name: "legacy", target: HookTarget{Target: "filter/tool"}, want: `"filter/tool"`},
		{name: "legacy mutation disabled", target: HookTarget{Target: "filter/tool", MutateDisallowed: true}, want: `"!mutate:filter/tool"`},
		{name: "v1", target: HookTarget{Target: "filter/tool", ContractVersion: filtercontract.ContractVersionV1}, want: `"!filter-v1:filter/tool"`},
		{name: "v1 mutation disabled", target: HookTarget{Target: "filter/tool", MutateDisallowed: true, ContractVersion: filtercontract.ContractVersionV1}, want: `"!mutate:!filter-v1:filter/tool"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.target)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != tt.want {
				t.Fatalf("got %s, want %s", data, tt.want)
			}
			var roundTrip HookTarget
			if err := json.Unmarshal(data, &roundTrip); err != nil {
				t.Fatal(err)
			}
			if roundTrip != tt.target {
				t.Fatalf("round trip = %#v, want %#v", roundTrip, tt.target)
			}
		})
	}
}

func TestLegacyHooksJSONIsUnchanged(t *testing.T) {
	hooks := Hooks{{
		Name:   "tools/call",
		Params: map[string]string{"name": "search"},
		Targets: []HookTarget{
			{Target: "first/filter"},
			{Target: "second/filter", MutateDisallowed: true},
		},
	}}
	data, err := json.Marshal(hooks)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"tools/call?name=search":["first/filter","!mutate:second/filter"]}`
	if string(data) != want {
		t.Fatalf("got %s, want %s", data, want)
	}
}
