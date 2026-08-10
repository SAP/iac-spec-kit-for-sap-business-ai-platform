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
}

// KnownAgents is the registry of supported agents, keyed by ID.
var KnownAgents = map[string]Agent{
	"claude":  {ID: "claude", Dir: filepath.Join(".claude", "commands", "btp-iac"), Ext: ".md"},
	"cursor":  {ID: "cursor", Dir: filepath.Join(".cursor", "rules"), Ext: ".mdc"},
	"copilot": {ID: "copilot", Dir: filepath.Join(".github", "instructions"), Ext: ".instructions.md"},
}

var baseDirs = []string{"specs", "memory", "terraform"}

const gitignore = `.terraform/
*.tfstate
*.tfstate.backup
.terraform.lock.hcl
`

// Scaffold creates the project directory named name in the current working
// directory, installs command files for each selected agent, writes .gitignore,
// and runs git init.
func Scaffold(name string, commands embed.FS, agents []Agent) error {
	if _, err := os.Stat(name); err == nil {
		return fmt.Errorf("directory %q already exists", name)
	}

	if err := os.Mkdir(name, 0o755); err != nil {
		return fmt.Errorf("create project directory: %w", err)
	}

	if err := scaffold(name, commands, agents); err != nil {
		_ = os.RemoveAll(name)
		return err
	}
	return nil
}

func scaffold(name string, commands embed.FS, agents []Agent) error {
	for _, d := range baseDirs {
		if err := os.MkdirAll(filepath.Join(name, d), 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", d, err)
		}
	}

	for _, agent := range agents {
		if err := os.MkdirAll(filepath.Join(name, agent.Dir), 0o755); err != nil {
			return fmt.Errorf("create agent directory %s: %w", agent.Dir, err)
		}
		if err := copyCommandsForAgent(name, commands, agent); err != nil {
			return err
		}
	}

	if err := os.WriteFile(filepath.Join(name, ".gitignore"), []byte(gitignore), 0o644); err != nil {
		return fmt.Errorf("write .gitignore: %w", err)
	}

	cmd := exec.Command("git", "init", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git init: %w", err)
	}

	return nil
}

func copyCommandsForAgent(projectDir string, commands embed.FS, agent Agent) error {
	destDir := filepath.Join(projectDir, agent.Dir)
	return fs.WalkDir(commands, "commands", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := commands.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded file %s: %w", path, err)
		}
		// Strip the source .md extension and apply the agent's extension.
		base := strings.TrimSuffix(filepath.Base(path), ".md")
		dest := filepath.Join(destDir, base+agent.Ext)
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", dest, err)
		}
		return nil
	})
}
