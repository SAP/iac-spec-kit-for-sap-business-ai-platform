package scaffold

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/SAP/btp-iac-spec-kit/skills"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

func TestGeminiInstructionsAndTOMLRoundTrip(t *testing.T) {
	body := "\n# Review \"quotes\" and \"\"\" delimiters\nPath: C:\\work\\project — 日本語\n" +
		"Prefer the sap-docs MCP server (tools prefixed `mcp__sap-docs__*`), then `WebFetch`.\n"
	metadata := "---\nname: sap-iac-fixture\ndescription: Review special characters\nlicense: Apache-2.0\nmetadata:\n  author: SAP\n---\n"
	skill, command, err := geminiInstructions([]byte(metadata+body), "sap-iac-fixture")
	if err != nil {
		t.Fatal(err)
	}
	want := "\n# Review \"quotes\" and \"\"\" delimiters\nPath: C:\\work\\project — 日本語\n" +
		"Prefer the sap-docs MCP server (tools exposed by the configured `sap-docs` MCP server), then `web_fetch`.\n"
	if command.Prompt != want || string(skill) != metadata+want || command.Description != "Review special characters" {
		t.Fatal("Gemini adaptation changed metadata or unexpected instruction content")
	}
	encoded, err := toml.Marshal(command)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]string
	if err := toml.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["prompt"] != want || decoded["description"] != command.Description {
		t.Fatal("TOML round trip altered prompt or description")
	}
}

func TestGeminiRejectsInvalidSkillMetadata(t *testing.T) {
	for _, source := range []string{
		"# No metadata",
		"---\nname: sap-iac-fixture\n",
		"---\nname: wrong-name\ndescription: test\n---\nbody",
		"---\nname: sap-iac-fixture\ndescription: \"\"\n---\nbody",
		"---\nname: [\n---\nbody",
		"---\nname: sap-iac-fixture\ndescription: test\n---\n",
	} {
		if _, _, err := geminiInstructions([]byte(source), "sap-iac-fixture"); err == nil {
			t.Errorf("accepted invalid skill: %q", source)
		}
	}
}

func TestGeminiOutputsAcrossModes(t *testing.T) {
	for _, mode := range []Mode{ModeFresh, ModeAdopt, ModeAgentOnly} {
		t.Run(map[Mode]string{ModeFresh: "fresh", ModeAdopt: "adopt", ModeAgentOnly: "agent-only"}[mode], func(t *testing.T) {
			dir := t.TempDir()
			if _, err := Apply(dir, mode, skills.Commands, []Agent{KnownAgents["gemini"]}); err != nil {
				t.Fatal(err)
			}
			commands, err := os.ReadDir(filepath.Join(dir, ".gemini", "commands"))
			if err != nil || len(commands) != 11 {
				t.Fatalf("command set: %v, %v", commands, err)
			}
			installedSkills, err := os.ReadDir(filepath.Join(dir, ".gemini", "skills"))
			if err != nil || len(installedSkills) != 11 {
				t.Fatalf("skill set: %v, %v", installedSkills, err)
			}
			for _, name := range []string{"govern", "scenario", "analyse", "accounts", "services", "security", "connectivity", "tasks", "design", "generate", "next"} {
				commandData := readGeminiTestFile(t, dir, ".gemini/commands/sap-iac."+name+".toml")
				var command map[string]string
				if err := toml.Unmarshal(commandData, &command); err != nil {
					t.Fatalf("%s: %v", name, err)
				}
				skillName := "sap-iac-" + name
				installed := readGeminiTestFile(t, dir, ".gemini/skills/"+skillName+"/SKILL.md")
				metadata, body, ok := strings.Cut(strings.TrimPrefix(string(installed), "---\n"), "\n---\n")
				var fields map[string]any
				if !ok || yaml.Unmarshal([]byte(metadata), &fields) != nil {
					t.Fatalf("%s: invalid native metadata", name)
				}
				if fields["name"] != skillName || fields["description"] != command["description"] || body != command["prompt"] || len(body) < 100 {
					t.Fatalf("%s: incomplete or inconsistent Gemini outputs", name)
				}
				source, err := skills.Commands.ReadFile(skillName + "/SKILL.md")
				if err != nil {
					t.Fatal(err)
				}
				_, originalBody, _ := strings.Cut(string(source[4:]), "\n---\n")
				want := strings.ReplaceAll(originalBody, "WebFetch", "web_fetch")
				want = strings.ReplaceAll(want, "tools prefixed `mcp__sap-docs__*`", "tools exposed by the configured `sap-docs` MCP server")
				if body != want {
					t.Fatalf("%s: substantive source instructions lost", name)
				}
			}
			catalogue, err := skills.Commands.ReadFile("sap-iac-services/" + CatalogueFile)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(catalogue, readGeminiTestFile(t, dir, ".gemini/skills/sap-iac-services/"+CatalogueFile)) {
				t.Fatal("bundled catalogue changed")
			}
			if err := filepath.WalkDir(filepath.Join(dir, ".gemini"), func(path string, entry fs.DirEntry, err error) error {
				if err == nil && entry.Name() == "evals" {
					t.Errorf("source eval assets installed at %s", path)
				}
				return err
			}); err != nil {
				t.Fatal(err)
			}
			for _, other := range []string{".claude", ".codex", ".cursor", ".github"} {
				if _, err := os.Stat(filepath.Join(dir, other)); !os.IsNotExist(err) {
					t.Errorf("unexpected agent directory %s", other)
				}
			}
			if mode == ModeAgentOnly {
				if _, err := os.Stat(filepath.Join(dir, "memory")); !os.IsNotExist(err) {
					t.Fatal("agent-only installation created memory content")
				}
			} else {
				for _, base := range baseDirs {
					if _, err := os.Stat(filepath.Join(dir, base)); err != nil {
						t.Errorf("missing %s: %v", base, err)
					}
				}
			}
		})
	}
}

func TestGeminiRefreshPreservesUnrelatedContent(t *testing.T) {
	dir := t.TempDir()
	preserved := []string{"specs/scenario.md", "memory/" + CatalogueFile, "memory/global-account.md", "terraform/main.tf", ".gemini/settings.json", ".gemini/commands/custom.toml", ".gemini/skills/custom/SKILL.md"}
	for _, rel := range preserved {
		if err := writeGeminiFile(filepath.Join(dir, filepath.FromSlash(rel)), []byte("user content")); err != nil {
			t.Fatal(err)
		}
	}
	agents := []Agent{KnownAgents["claude"], KnownAgents["gemini"]}
	if _, err := Apply(dir, ModeAdopt, skills.Commands, agents); err != nil {
		t.Fatal(err)
	}
	claudePath := ".claude/commands/sap-iac.scenario.md"
	claude := readGeminiTestFile(t, dir, claudePath)
	source, err := skills.Commands.ReadFile("sap-iac-scenario/SKILL.md")
	if err != nil || !bytes.Equal(claude, source) {
		t.Fatal("mixed installation changed Claude source content")
	}
	managed := []string{".gemini/commands/sap-iac.scenario.toml", ".gemini/skills/sap-iac-scenario/SKILL.md", ".gemini/skills/sap-iac-services/" + CatalogueFile}
	expected := make(map[string][]byte)
	for _, rel := range managed {
		expected[rel] = readGeminiTestFile(t, dir, rel)
		if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(rel)), []byte("stale content"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Apply(dir, ModeAgentOnly, skills.Commands, []Agent{KnownAgents["gemini"]}); err != nil {
		t.Fatal(err)
	}
	for rel, want := range expected {
		if !bytes.Equal(readGeminiTestFile(t, dir, rel), want) {
			t.Errorf("managed output not refreshed: %s", rel)
		}
	}
	for _, rel := range preserved {
		if string(readGeminiTestFile(t, dir, rel)) != "user content" {
			t.Errorf("user content changed: %s", rel)
		}
	}
	if !bytes.Equal(readGeminiTestFile(t, dir, claudePath), claude) {
		t.Fatal("Gemini refresh changed unselected Claude files")
	}
}

func TestGeminiNestedRuntimeResources(t *testing.T) {
	assets := fstest.MapFS{
		"sap-iac-fixture/SKILL.md":                {Data: []byte("---\nname: sap-iac-fixture\ndescription: Fixture\n---\n# Instructions\n")},
		"sap-iac-fixture/references/example.yaml": {Data: []byte("nested resource")},
	}
	dir := t.TempDir()
	if err := installGemini(dir, assets, KnownAgents["gemini"]); err != nil {
		t.Fatal(err)
	}
	if string(readGeminiTestFile(t, dir, ".gemini/skills/sap-iac-fixture/references/example.yaml")) != "nested resource" {
		t.Fatal("resource path or content changed")
	}
}

func readGeminiTestFile(t *testing.T, dir, rel string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return data
}
