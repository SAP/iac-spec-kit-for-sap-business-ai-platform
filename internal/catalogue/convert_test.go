package catalogue

import (
	"strings"
	"testing"
)

func TestConvert_FlatProperties(t *testing.T) {
	schema := `{
		"type": "object",
		"required": ["appname"],
		"properties": {
			"appname": {
				"type": "string",
				"description": "Application name"
			},
			"mode": {
				"type": "string",
				"enum": ["dedicated", "shared"],
				"description": "Tenancy mode"
			},
			"count": {
				"type": "integer",
				"default": 3
			},
			"notype": {}
		}
	}`

	out, err := Convert([]byte(schema), "my-service", []string{"standard"})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	s := string(out)

	// source must be user-defined (task 3.5)
	if !strings.Contains(s, "source: user-defined") {
		t.Error("missing source: user-defined")
	}
	// required property must appear under required
	if !strings.Contains(s, "key: appname") {
		t.Error("required property appname missing")
	}
	// optional properties present
	if !strings.Contains(s, "key: mode") {
		t.Error("optional property mode missing")
	}
	// description copied
	if !strings.Contains(s, "Application name") {
		t.Error("description not copied")
	}
	// enum synthesised when no description
	if !strings.Contains(s, "dedicated") {
		t.Error("enum values missing from description or example")
	}
	// default used as example
	if !strings.Contains(s, "3") {
		t.Error("default value not used as example")
	}
}

func TestConvert_RequiredVsOptional(t *testing.T) {
	schema := `{
		"type": "object",
		"required": ["req1", "req2"],
		"properties": {
			"req1": {"type": "string"},
			"req2": {"type": "string"},
			"opt1": {"type": "string"}
		}
	}`
	out, err := Convert([]byte(schema), "svc", []string{"plan"})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	s := string(out)

	reqIdx := strings.Index(s, "required:")
	optIdx := strings.Index(s, "optional:")
	if reqIdx < 0 || optIdx < 0 {
		t.Fatal("required or optional section missing")
	}
	req1Idx := strings.Index(s, "req1")
	req2Idx := strings.Index(s, "req2")
	opt1Idx := strings.Index(s, "opt1")
	if req1Idx < reqIdx || req2Idx < reqIdx {
		t.Error("required properties not under required section")
	}
	if opt1Idx < optIdx {
		t.Error("optional property not under optional section")
	}
}

func TestConvert_NestedObject(t *testing.T) {
	schema := `{
		"type": "object",
		"properties": {
			"config": {
				"type": "object",
				"description": "Config block",
				"properties": {
					"enabled": {"type": "boolean", "default": false},
					"limit":   {"type": "integer", "default": 100}
				}
			}
		}
	}`
	out, err := Convert([]byte(schema), "svc", []string{"plan"})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	s := string(out)
	// nested object must produce a map, not a plain string placeholder
	if !strings.Contains(s, "enabled:") || !strings.Contains(s, "limit:") {
		t.Errorf("nested object not expanded into YAML map: %s", s)
	}
	// must not be a plain string example
	if strings.Contains(s, "example: <config>") {
		t.Error("nested object produced string placeholder instead of map")
	}
}

func TestConvert_OneOf(t *testing.T) {
	schema := `{
		"type": "object",
		"properties": {
			"auth": {
				"oneOf": [
					{"type": "string", "description": "Token auth"},
					{"type": "object", "description": "Cert auth"}
				]
			}
		}
	}`
	out, err := Convert([]byte(schema), "svc", []string{"plan"})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "simplified from oneOf") {
		t.Error("oneOf comment missing")
	}
	// first branch description used
	if !strings.Contains(s, "Token auth") {
		t.Error("first oneOf branch not used")
	}
}

func TestConvert_AnyOf(t *testing.T) {
	schema := `{
		"type": "object",
		"properties": {
			"value": {
				"anyOf": [
					{"type": "string", "description": "String form"},
					{"type": "integer"}
				]
			}
		}
	}`
	out, err := Convert([]byte(schema), "svc", []string{"plan"})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "simplified from anyOf") {
		t.Error("anyOf comment missing")
	}
}

func TestConvert_InternalRef(t *testing.T) {
	schema := `{
		"type": "object",
		"properties": {
			"version": {"$ref": "#/$defs/Version"}
		},
		"$defs": {
			"Version": {
				"type": "object",
				"description": "Version definition",
				"properties": {
					"major": {"type": "integer", "default": 1}
				}
			}
		}
	}`
	out, err := Convert([]byte(schema), "svc", []string{"plan"})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "Version definition") {
		t.Errorf("internal $ref not resolved: %s", s)
	}
	if !strings.Contains(s, "major:") {
		t.Errorf("nested properties of $ref not expanded: %s", s)
	}
}

func TestConvert_InternalRefMissingKeyProducesPlaceholder(t *testing.T) {
	schema := `{
		"type": "object",
		"properties": {
			"cfg": {"$ref": "#/$defs/Missing"}
		},
		"$defs": {}
	}`
	out, err := Convert([]byte(schema), "svc", []string{"plan"})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "unresolved $ref") {
		t.Errorf("missing internal $ref should produce unresolved placeholder: %s", s)
	}
	// must not look like a normal placeholder
	if strings.Contains(s, "example: <cfg>") {
		t.Errorf("missing internal $ref produced generic placeholder instead of ref placeholder: %s", s)
	}
}

func TestConvert_RootOneOfWithDefsOnOriginalDocument(t *testing.T) {
	// The root oneOf branch references $defs that live on the outer document,
	// not inside the branch. resolveRef must use the original root, not the branch.
	schema := `{
		"oneOf": [
			{
				"type": "object",
				"properties": {
					"cfg": {"$ref": "#/$defs/Config"}
				}
			}
		],
		"$defs": {
			"Config": {
				"type": "object",
				"description": "Config definition",
				"properties": {
					"level": {"type": "integer", "default": 1}
				}
			}
		}
	}`
	out, err := Convert([]byte(schema), "svc", []string{"plan"})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	s := string(out)
	if strings.Contains(s, "unresolved $ref") {
		t.Errorf("$ref to $defs on original document should resolve, got unresolved placeholder:\n%s", s)
	}
	if !strings.Contains(s, "Config definition") {
		t.Errorf("resolved $ref description missing:\n%s", s)
	}
	if !strings.Contains(s, "level:") {
		t.Errorf("nested property from resolved $ref missing:\n%s", s)
	}
}

func TestConvert_ExternalRefProducesPlaceholder(t *testing.T) {
	schema := `{
		"type": "object",
		"properties": {
			"remote": {"$ref": "https://example.com/schema.json#/Foo"}
		}
	}`
	out, err := Convert([]byte(schema), "svc", []string{"plan"})
	if err != nil {
		t.Fatalf("Convert must not error on external ref: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "unresolved $ref") {
		t.Errorf("external $ref should produce placeholder: %s", s)
	}
}

func TestConvert_RootLevelOneOf(t *testing.T) {
	schema := `{
		"oneOf": [
			{
				"type": "object",
				"required": ["name"],
				"properties": {
					"name": {"type": "string", "description": "Instance name"},
					"count": {"type": "integer", "default": 1}
				}
			},
			{"type": "string"}
		]
	}`
	out, err := Convert([]byte(schema), "svc", []string{"plan"})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	s := string(out)
	if strings.Contains(s, "required:\n    - []\n") || !strings.Contains(s, "key: name") {
		t.Errorf("root-level oneOf not resolved: properties empty or name missing:\n%s", s)
	}
	if !strings.Contains(s, "key: count") {
		t.Errorf("optional property from root oneOf branch missing:\n%s", s)
	}
}

func TestConvert_ExternalRefComment(t *testing.T) {
	schema := `{
		"type": "object",
		"properties": {
			"cert": {"$ref": "https://example.com/cert.json"}
		}
	}`
	out, err := Convert([]byte(schema), "svc", []string{"plan"})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	s := string(out)
	// description must contain the placeholder text, not just "<cert>"
	if !strings.Contains(s, "<unresolved $ref: https://example.com/cert.json>") {
		t.Errorf("unresolved $ref example should be the ref URL, got:\n%s", s)
	}
	// YAML comment must be present
	if !strings.Contains(s, "unresolved $ref:") {
		t.Errorf("unresolved $ref comment missing:\n%s", s)
	}
}

func TestConvert_SourceIsUserDefined(t *testing.T) {
	schema := `{"type":"object","properties":{"x":{"type":"string"}}}`
	out, err := Convert([]byte(schema), "svc", []string{"plan"})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if !strings.Contains(string(out), "source: user-defined") {
		t.Error("source must be user-defined")
	}
}

func TestConvert_BTPServiceCatalog(t *testing.T) {
	// BTP Open Service Broker API shape: schemas is map[string]any, not []any.
	catalog := `{
		"servicePlans": [{
			"name": "standard",
			"schemas": {
				"service_instance": {
					"create": {
						"parameters": {
							"type": "object",
							"required": ["region"],
							"properties": {
								"region": {
									"type": "string",
									"description": "Deployment region"
								},
								"tier": {
									"type": "string",
									"description": "Service tier"
								}
							}
						}
					}
				}
			}
		}]
	}`
	out, err := Convert([]byte(catalog), "my-btp-service", []string{"standard"})
	if err != nil {
		t.Fatalf("Convert BTP catalog: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "service: my-btp-service") {
		t.Errorf("service name missing: %s", s)
	}
	if !strings.Contains(s, "Deployment region") {
		t.Errorf("required param description missing: %s", s)
	}
	if !strings.Contains(s, "key: region") {
		t.Errorf("required param key missing: %s", s)
	}
	if !strings.Contains(s, "key: tier") {
		t.Errorf("optional param key missing: %s", s)
	}
}

func TestConvert_BTPServiceCatalogMissingSchemas(t *testing.T) {
	catalog := `{"servicePlans": [{"name": "standard"}]}`
	_, err := Convert([]byte(catalog), "svc", []string{"standard"})
	if err == nil {
		t.Fatal("expected error for missing schemas")
	}
}

func TestConvert_PlansInOutput(t *testing.T) {
	schema := `{"type":"object","properties":{}}`
	out, err := Convert([]byte(schema), "my-service", []string{"plan-a", "plan-b"})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "plan-a") || !strings.Contains(s, "plan-b") {
		t.Errorf("plans missing from output: %s", s)
	}
	if !strings.Contains(s, "service: my-service") {
		t.Errorf("service name missing: %s", s)
	}
}
