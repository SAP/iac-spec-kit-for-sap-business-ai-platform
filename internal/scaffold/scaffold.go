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

	"github.com/SAP/btp-iac-spec-kit/internal/platformvalidation"
)

// Mode controls which parts of the project are created or updated.
type Mode int

const (
	// ModeFresh creates a new project subdirectory with all dirs, agent files,
	// .gitignore, and runs git init.
	ModeFresh Mode = iota + 1
	// ModeAdopt sets up btp-iac inside an existing directory (e.g. a Terraform
	// repo). Creates specs/, memory/, terraform/, agent files, and .gitignore
	// if absent. Skips git init.
	ModeAdopt
	// ModeAgentOnly updates agent command files and ensures the local
	// platform-validation record is ignored, leaving project content untouched.
	ModeAgentOnly
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
	"codex":   {ID: "codex", Dir: filepath.Join(".codex", "prompts"), Ext: ".md", Prefix: "btp-iac."},
	"cursor":  {ID: "cursor", Dir: filepath.Join(".cursor", "rules"), Ext: ".mdc", Prefix: "btp-iac."},
	"copilot": {ID: "copilot", Dir: filepath.Join(".github", "instructions"), Ext: ".instructions.md", Prefix: "btp-iac."},
}

var baseDirs = []string{"specs", "memory", "terraform"}

// GlobalAccountFile is the project-local source of truth for the global
// account selected during initialization. It is deliberately separate from
// governance so every skill can use it without prompting again.
const GlobalAccountFile = "global-account.md"

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
	warning, err = Apply(name, ModeFresh, commands, agents)
	if err != nil {
		_ = os.RemoveAll(name)
		return "", err
	}
	return warning, nil
}

// IsProject reports whether dir looks like an existing btp-iac project by
// checking for the presence of specs/ or memory/.
func IsProject(dir string) bool {
	for _, marker := range []string{"specs", "memory"} {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			return true
		}
	}
	return false
}

// Apply runs the scaffold operation for the given mode inside dir.
// dir must already exist for ModeAdopt and ModeAgentOnly.
// warning is non-empty when git was not found (ModeFresh only).
func Apply(dir string, mode Mode, commands embed.FS, agents []Agent) (warning string, err error) {
	if mode == ModeFresh || mode == ModeAdopt {
		for _, d := range baseDirs {
			if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
				return "", fmt.Errorf("create directory %s: %w", d, err)
			}
		}
		if err := WriteCatalogueFile(dir, commands); err != nil {
			return "", err
		}
	}

	for _, agent := range agents {
		if err := os.MkdirAll(filepath.Join(dir, agent.Dir), 0o755); err != nil {
			return "", fmt.Errorf("create agent directory %s: %w", agent.Dir, err)
		}
		if err := copyCommandsForAgent(dir, commands, agent); err != nil {
			return "", err
		}
	}

	for _, entry := range []string{platformvalidation.Ignore, filepath.Join("memory", GlobalAccountFile)} {
		if err := ensureGitignoreEntry(dir, entry); err != nil {
			return "", err
		}
	}

	if mode != ModeFresh {
		return "", nil
	}

	if _, err := exec.LookPath("git"); err != nil {
		// ponytail: skip git init when git is absent; caller prints the warning
		return `"git" was not found on $PATH — run "git init" manually in the project directory.`, nil
	}

	cmd := exec.Command("git", "init", dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git init: %w", err)
	}

	return "", nil
}

// CatalogueFile is the name of the service parameters catalogue written to memory/.
const CatalogueFile = "service-params-catalogue.yaml"

// WriteCatalogueFile copies the embedded service-params-catalogue.yaml to
// <dir>/memory/service-params-catalogue.yaml. It is a no-op when the file
// already exists, preserving any user edits on re-init.
func WriteCatalogueFile(dir string, commands embed.FS) error {
	dest := filepath.Join(dir, "memory", CatalogueFile)
	if _, err := os.Stat(dest); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat catalogue file: %w", err)
	}
	data, err := commands.ReadFile("btp-iac-services/" + CatalogueFile)
	if err != nil {
		return fmt.Errorf("read embedded catalogue: %w", err)
	}
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return fmt.Errorf("write catalogue file: %w", err)
	}
	return nil
}

// WriteGlobalAccountSubdomain stores the optional global-account subdomain
// selected during initialization. An empty input creates the record for a new
// project but never replaces an existing configured value.
func WriteGlobalAccountSubdomain(dir, subdomain string) error {
	path := filepath.Join(dir, "memory", GlobalAccountFile)
	if subdomain == "" {
		if _, err := os.Stat(path); err == nil {
			return nil
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("stat global account record: %w", err)
		}
	}

	content := "# Global Account\n\n- Subdomain:\n"
	if subdomain != "" {
		content = fmt.Sprintf("# Global Account\n\n- Subdomain: %s\n", subdomain)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write global account record: %w", err)
	}
	return nil
}

func ensureGitignoreEntry(dir, entry string) error {
	path := filepath.Join(dir, ".gitignore")
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read .gitignore: %w", err)
	}
	if os.IsNotExist(err) {
		data = append([]byte(gitignore), entry+"\n"...)
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return fmt.Errorf("write .gitignore: %w", err)
		}
		return nil
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if line == entry ||
			line == platformvalidation.Directory ||
			line == platformvalidation.Directory+"/" ||
			(entry == filepath.Join("memory", GlobalAccountFile) && (line == "memory" || line == "memory/")) {
			return nil
		}
	}
	if len(data) > 0 && !strings.HasSuffix(string(data), "\n") {
		data = append(data, '\n')
	}
	data = append(data, entry+"\n"...)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write .gitignore: %w", err)
	}
	return nil
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
