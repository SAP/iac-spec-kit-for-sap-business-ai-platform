// Package catalogue converts JSON Schema documents into sap-iac service
// parameter catalogue YAML entries.
package catalogue

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Entry is the top-level catalogue entry written to YAML.
type Entry struct {
	Service    string     `yaml:"service"`
	Plans      []string   `yaml:"plans"`
	Source     string     `yaml:"source"`
	Parameters ParamGroup `yaml:"parameters"`
}

// ParamGroup holds required and optional parameter definitions.
type ParamGroup struct {
	Required []Param `yaml:"required"`
	Optional []Param `yaml:"optional"`
}

// Param is a single parameter definition.
type Param struct {
	Key         string `yaml:"key"`
	Description string `yaml:"description"`
	Example     any    `yaml:"example"`
	// comment is emitted as a YAML line comment when non-empty.
	comment string
}

// Convert parses schemaBytes as a JSON Schema (or BTP service catalog entry)
// and returns a YAML catalogue entry for the given service and plans.
// The entry carries source: user-defined.
//
// When the input is a BTP service catalog entry (contains "servicePlans"),
// Convert extracts the first plan's service_instance create parameters schema.
//
// Conversion rules:
//   - properties listed in required[] → parameters.required
//   - all other properties → parameters.optional
//   - description copied from schema; synthesized from type/enum when absent
//   - example: default > enum[0] > "<key-name>"
//   - nested "type":"object" → nested YAML map in example
//   - oneOf/anyOf → first branch used, comment added
//   - internal $ref → resolved from $defs or definitions
func Convert(schemaBytes []byte, service string, plans []string) ([]byte, error) {
	var root map[string]any
	if err := json.Unmarshal(schemaBytes, &root); err != nil {
		return nil, fmt.Errorf("parse schema: %w", err)
	}

	// Auto-detect BTP service catalog wrapper and extract the actual schema.
	if sp, ok := root["servicePlans"].([]any); ok && len(sp) > 0 {
		extracted, err := extractSchemaFromCatalog(root)
		if err != nil {
			return nil, err
		}
		root = extracted
	}

	// Root-level oneOf/anyOf: pick the first branch as the active schema for
	// property enumeration, but keep root as the $ref resolver so that $defs
	// and definitions on the original document remain reachable.
	active := root
	if branch := firstBranch(root, "oneOf"); branch != nil {
		active = resolveRef(branch, root)
	} else if branch := firstBranch(root, "anyOf"); branch != nil {
		active = resolveRef(branch, root)
	}

	requiredSet := stringSet(jsonStringSlice(active["required"]))
	props, _ := active["properties"].(map[string]any)

	entry := Entry{
		Service: service,
		Plans:   plans,
		Source:  "user-defined",
		Parameters: ParamGroup{
			Required: []Param{},
			Optional: []Param{},
		},
	}

	for key, val := range props {
		prop, _ := val.(map[string]any)
		prop = resolveRef(prop, root)
		p := buildParam(key, prop, root)
		if requiredSet[key] {
			entry.Parameters.Required = append(entry.Parameters.Required, p)
		} else {
			entry.Parameters.Optional = append(entry.Parameters.Optional, p)
		}
	}

	return marshalEntry(entry)
}

// extractSchemaFromCatalog pulls the service_instance create parameters schema
// from a BTP service catalog entry (the first plan's first schema entry).
func extractSchemaFromCatalog(catalog map[string]any) (map[string]any, error) {
	sp, _ := catalog["servicePlans"].([]any)
	if len(sp) == 0 {
		return nil, fmt.Errorf("servicePlans is empty")
	}
	plan, _ := sp[0].(map[string]any)
	schemas, _ := plan["schemas"].([]any)
	if len(schemas) == 0 {
		return nil, fmt.Errorf("no schemas in first service plan")
	}
	schemaEntry, _ := schemas[0].(map[string]any)
	si, _ := schemaEntry["service_instance"].(map[string]any)
	create, _ := si["create"].(map[string]any)
	params, _ := create["parameters"].(map[string]any)
	if params == nil {
		return nil, fmt.Errorf("no service_instance.create.parameters in catalog entry")
	}
	return params, nil
}

// buildParam constructs a Param from a JSON Schema property definition.
func buildParam(key string, prop map[string]any, root map[string]any) Param {
	p := Param{Key: key}

	// Handle oneOf / anyOf first — replace prop with the first branch before
	// extracting description and example.
	comment := ""
	if branches := firstBranch(prop, "oneOf"); branches != nil {
		prop = resolveRef(branches, root)
		comment = "simplified from oneOf"
	} else if branches := firstBranch(prop, "anyOf"); branches != nil {
		prop = resolveRef(branches, root)
		comment = "simplified from anyOf"
	}

	// Unresolved $ref: mark with a comment so the YAML is clearly a placeholder.
	if ref, ok := prop["_unresolved"].(string); ok && ref != "" {
		comment = "unresolved $ref: " + ref
	}
	p.comment = comment

	// Description
	if d, ok := prop["description"].(string); ok && d != "" {
		p.Description = d
	} else if enum := jsonStringSlice(prop["enum"]); len(enum) > 0 {
		p.Description = "Allowed values: " + strings.Join(enum, ", ")
	} else if t, ok := prop["type"].(string); ok {
		p.Description = "Type: " + t
	} else {
		p.Description = key
	}

	// Example
	p.Example = exampleValue(key, prop, root)

	return p
}

// exampleValue derives an example value from a property schema.
func exampleValue(key string, prop map[string]any, root map[string]any) any {
	// Unresolved $ref: use the ref string as the placeholder so the YAML is
	// clearly not a real value.
	if ref, ok := prop["_unresolved"].(string); ok && ref != "" {
		return "<unresolved $ref: " + ref + ">"
	}
	// default value wins
	if def, ok := prop["default"]; ok {
		return def
	}
	// enum: use first value
	if enum := jsonAnySlice(prop["enum"]); len(enum) > 0 {
		return enum[0]
	}
	// nested object: recurse into properties
	if t, _ := prop["type"].(string); t == "object" {
		if subProps, ok := prop["properties"].(map[string]any); ok {
			nested := map[string]any{}
			for subKey, subVal := range subProps {
				subProp, _ := subVal.(map[string]any)
				subProp = resolveRef(subProp, root)
				nested[subKey] = exampleValue(subKey, subProp, root)
			}
			return nested
		}
		return map[string]any{}
	}
	// array: emit a one-element example
	if t, _ := prop["type"].(string); t == "array" {
		return []any{"<" + key + "-item>"}
	}
	return "<" + key + ">"
}

// resolveRef follows a $ref within the same document ($defs or definitions).
// Returns the original map unchanged when no $ref is present or it cannot
// be resolved (external refs, missing keys).
func resolveRef(prop map[string]any, root map[string]any) map[string]any {
	ref, ok := prop["$ref"].(string)
	if !ok {
		return prop
	}
	// Only handle internal refs: #/$defs/Foo or #/definitions/Foo
	const internalPrefix = "#/"
	if !strings.HasPrefix(ref, internalPrefix) {
		return map[string]any{
			"description": "<unresolved $ref: " + ref + ">",
			"_unresolved": ref,
		}
	}
	parts := strings.Split(strings.TrimPrefix(ref, internalPrefix), "/")
	node := map[string]any(root)
	for _, part := range parts {
		child, ok := node[part].(map[string]any)
		if !ok {
			return map[string]any{
				"description": "<unresolved $ref: " + ref + ">",
				"_unresolved": ref,
			}
		}
		node = child
	}
	return node
}

// firstBranch returns the first entry of a oneOf/anyOf array, or nil.
func firstBranch(prop map[string]any, key string) map[string]any {
	arr, ok := prop[key].([]any)
	if !ok || len(arr) == 0 {
		return nil
	}
	branch, _ := arr[0].(map[string]any)
	return branch
}

// marshalEntry serialises an Entry to YAML, adding line comments where needed.
func marshalEntry(e Entry) ([]byte, error) {
	// Build a yaml.Node tree so we can attach comments.
	doc := &yaml.Node{Kind: yaml.DocumentNode}
	seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	doc.Content = append(doc.Content, seq)

	entryMap := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	seq.Content = append(seq.Content, entryMap)

	appendStr(entryMap, "service", e.Service)
	appendStrSeq(entryMap, "plans", e.Plans)
	appendStr(entryMap, "source", e.Source)

	// parameters map
	appendStr(entryMap, "parameters", "")
	paramsKey := entryMap.Content[len(entryMap.Content)-2]
	paramsKey.Value = "parameters"
	paramsMap := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	entryMap.Content[len(entryMap.Content)-1] = paramsMap

	appendParamList(paramsMap, "required", e.Parameters.Required)
	appendParamList(paramsMap, "optional", e.Parameters.Optional)

	out, err := yaml.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("marshal entry: %w", err)
	}
	return out, nil
}

func appendParamList(parent *yaml.Node, key string, params []Param) {
	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key, Tag: "!!str"}
	seqNode := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	parent.Content = append(parent.Content, keyNode, seqNode)

	for _, p := range params {
		item := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		appendStr(item, "key", p.Key)
		descNode := &yaml.Node{Kind: yaml.ScalarNode, Value: p.Description, Tag: "!!str"}
		if p.comment != "" {
			descNode.LineComment = "# " + p.comment
		}
		item.Content = append(item.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "description", Tag: "!!str"},
			descNode,
		)
		exNode, _ := yaml.Marshal(p.Example)
		var exYAML yaml.Node
		if err := yaml.Unmarshal(exNode, &exYAML); err == nil && len(exYAML.Content) > 0 {
			item.Content = append(item.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "example", Tag: "!!str"},
				exYAML.Content[0],
			)
		}
		seqNode.Content = append(seqNode.Content, item)
	}
}

func appendStr(parent *yaml.Node, key, val string) {
	parent.Content = append(parent.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key, Tag: "!!str"},
		&yaml.Node{Kind: yaml.ScalarNode, Value: val, Tag: "!!str"},
	)
}

func appendStrSeq(parent *yaml.Node, key string, vals []string) {
	seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Style: yaml.FlowStyle}
	for _, v := range vals {
		seq.Content = append(seq.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: v, Tag: "!!str"})
	}
	parent.Content = append(parent.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key, Tag: "!!str"},
		seq,
	)
}

func stringSet(ss []string) map[string]bool {
	m := make(map[string]bool, len(ss))
	for _, s := range ss {
		m[s] = true
	}
	return m
}

func jsonStringSlice(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func jsonAnySlice(v any) []any {
	arr, _ := v.([]any)
	return arr
}
