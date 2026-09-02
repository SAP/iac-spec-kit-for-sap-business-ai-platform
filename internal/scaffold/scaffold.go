// Package scaffold creates the btp-iac project directory structure.
package scaffold

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Agent describes a supported AI coding agent and where its command files live.
type Agent struct {
	// ID is the canonical identifier used in the --agent flag (e.g. "claude").
	ID string
	// Dir is the destination directory relative to the project root.
	Dir string
	// Ext is the file extension written for command files (e.g. ".md", ".mdc").
	Ext string
	// Prefix is prepended to each command filename (e.g. "btp-iac.").
	Prefix string
}

// KnownAgents is the registry of supported agents, keyed by ID.
var KnownAgents = map[string]Agent{
	"claude":  {ID: "claude", Dir: filepath.Join(".claude", "commands"), Ext: ".md", Prefix: "btp-iac."},
	"cursor":  {ID: "cursor", Dir: filepath.Join(".cursor", "rules"), Ext: ".mdc", Prefix: "btp-iac."},
	"copilot": {ID: "copilot", Dir: filepath.Join(".github", "instructions"), Ext: ".instructions.md", Prefix: "btp-iac."},
}

var baseDirs = []string{"specs", "memory", "terraform"}

const gitignore = `.terraform/
*.tfstate
*.tfstate.backup
.terraform.lock.hcl
`

// Scaffold creates the project directory named name in the current working
// directory, installs command files for each selected agent, writes .gitignore,
// and runs git init if git is available.
// warning is non-empty when git was not found and git init was skipped.
func Scaffold(name string, commands embed.FS, agents []Agent) (warning string, err error) {
	if _, err := os.Stat(name); err == nil {
		return "", fmt.Errorf("directory %q already exists", name)
	}

	if err := os.Mkdir(name, 0o755); err != nil {
		return "", fmt.Errorf("create project directory: %w", err)
	}

	warning, err = scaffold(name, commands, agents)
	if err != nil {
		_ = os.RemoveAll(name)
		return "", err
	}
	return warning, nil
}

func scaffold(name string, commands embed.FS, agents []Agent) (string, error) {
	for _, d := range baseDirs {
		if err := os.MkdirAll(filepath.Join(name, d), 0o755); err != nil {
			return "", fmt.Errorf("create directory %s: %w", d, err)
		}
	}

	for _, agent := range agents {
		if err := os.MkdirAll(filepath.Join(name, agent.Dir), 0o755); err != nil {
			return "", fmt.Errorf("create agent directory %s: %w", agent.Dir, err)
		}
		if err := copyCommandsForAgent(name, commands, agent); err != nil {
			return "", err
		}
	}

	if err := os.WriteFile(filepath.Join(name, ".gitignore"), []byte(gitignore), 0o644); err != nil {
		return "", fmt.Errorf("write .gitignore: %w", err)
	}

	if _, err := exec.LookPath("git"); err != nil {
		// ponytail: skip git init when git is absent; caller prints the warning
		return `"git" was not found on $PATH — run "git init" manually in the project directory.`, nil
	}

	cmd := exec.Command("git", "init", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git init: %w", err)
	}

	return "", nil
}

func copyCommandsForAgent(projectDir string, commands embed.FS, agent Agent) error {
	destDir := filepath.Join(projectDir, agent.Dir)
	return fs.WalkDir(commands, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || path == "." {
			return nil
		}
		// Only process skill entry points; skip any stray files.
		if filepath.Base(path) != "SKILL.md" {
			return nil
		}
		data, err := commands.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded file %s: %w", path, err)
		}
		// Extract command name from the skill directory (e.g. "btp-iac-govern" → "govern").
		base := strings.TrimPrefix(filepath.Base(filepath.Dir(path)), "btp-iac-")
		dest := filepath.Join(destDir, agent.Prefix+base+agent.Ext)
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", dest, err)
		}
		return nil
	})
}
