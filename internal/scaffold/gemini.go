package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

type geminiCommand struct {
	Description string `toml:"description"`
	Prompt      string `toml:"prompt"`
}

// geminiInstructions retains the original metadata while adapting the body
// once for both the native skill and its explicit slash command.
func geminiInstructions(data []byte, skillName string) ([]byte, geminiCommand, error) {
	text := string(data)
	if !strings.HasPrefix(text, "---\n") {
		return nil, geminiCommand{}, fmt.Errorf("%s: missing skill frontmatter", skillName)
	}
	metadata, body, ok := strings.Cut(text[4:], "\n---\n")
	if !ok {
		return nil, geminiCommand{}, fmt.Errorf("%s: unterminated skill frontmatter", skillName)
	}
	var fields struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}
	if err := yaml.Unmarshal([]byte(metadata), &fields); err != nil {
		return nil, geminiCommand{}, fmt.Errorf("%s: parse skill metadata: %w", skillName, err)
	}
	if fields.Name != skillName || strings.TrimSpace(fields.Description) == "" || strings.TrimSpace(body) == "" {
		return nil, geminiCommand{}, fmt.Errorf("%s: invalid skill name, description, or body", skillName)
	}
	body = strings.NewReplacer(
		"WebFetch", "web_fetch",
		"tools prefixed `mcp__sap-docs__*`", "tools exposed by the configured `sap-docs` MCP server",
	).Replace(body)
	skill := []byte("---\n" + metadata + "\n---\n" + body)
	return skill, geminiCommand{Description: fields.Description, Prompt: body}, nil
}

// installGemini copies only embedded runtime assets, preserving skill-relative
// resource paths and leaving unrelated project commands and settings untouched.
func installGemini(projectDir string, commands fs.FS, agent Agent) error {
	entries, err := fs.ReadDir(commands, ".")
	if err != nil {
		return fmt.Errorf("list embedded skills: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "sap-iac-") {
			continue
		}
		skillName := entry.Name()
		data, err := fs.ReadFile(commands, path.Join(skillName, "SKILL.md"))
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", skillName, err)
		}
		skill, command, err := geminiInstructions(data, skillName)
		if err != nil {
			return err
		}
		commandData, err := toml.Marshal(command)
		if err != nil {
			return fmt.Errorf("encode Gemini command %s: %w", skillName, err)
		}
		commandName := commandFileName(agent, skillName)
		if err := writeGeminiFile(filepath.Join(projectDir, agent.Dir, commandName), commandData); err != nil {
			return err
		}
		skillDir := filepath.Join(projectDir, ".gemini", "skills", skillName)
		if err := fs.WalkDir(commands, skillName, func(assetPath string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				return nil
			}
			rel := strings.TrimPrefix(assetPath, skillName+"/")
			var asset []byte
			if rel == "SKILL.md" {
				asset = skill
			} else {
				asset, err = fs.ReadFile(commands, assetPath)
				if err != nil {
					return fmt.Errorf("read embedded resource %s: %w", assetPath, err)
				}
			}
			return writeGeminiFile(filepath.Join(skillDir, filepath.FromSlash(rel)), asset)
		}); err != nil {
			return err
		}
	}
	return nil
}

func writeGeminiFile(dest string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("create Gemini directory: %w", err)
	}
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", dest, err)
	}
	return nil
}
