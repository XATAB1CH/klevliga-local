package api

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// Catch malformed YAML, duplicate mapping keys, broken local references and
// repeated operation IDs without installing another schema tool in CI.
func TestOpenAPIReferences(t *testing.T) {
	raw, err := os.ReadFile("../../../../api/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := yaml.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	var walk func(any)
	walk = func(node any) {
		switch value := node.(type) {
		case map[string]any:
			if ref, ok := value["$ref"].(string); ok && strings.HasPrefix(ref, "#/") {
				var current any = document
				for _, part := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
					part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
					mapping, ok := current.(map[string]any)
					if !ok || mapping[part] == nil {
						t.Fatalf("unresolved reference: %s", ref)
					}
					current = mapping[part]
				}
			}
			if id, ok := value["operationId"].(string); ok {
				if ids[id] {
					t.Fatalf("duplicate operationId: %s", id)
				}
				ids[id] = true
			}
			for _, child := range value {
				walk(child)
			}
		case []any:
			for _, child := range value {
				walk(child)
			}
		}
	}
	walk(document)
	if len(ids) == 0 {
		t.Fatal("no API operations found")
	}
}
