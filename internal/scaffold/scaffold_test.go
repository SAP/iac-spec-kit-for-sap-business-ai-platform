package scaffold

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SAP/btp-iac-sdd/internal/assets"
)

func TestScaffoldClaude(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig) //nolint:errcheck

	agents := []Agent{KnownAgents["claude"]}
	if err := Scaffold("myproject", assets.Commands, agents); err != nil {
		t.Fatalf("Scaffold: %v", err)
	}

	expected := []string{
		"specs", "memory", "terraform",
		filepath.Join(".claude", "commands", "btp-iac"),
		".gitignore", ".git",
		filepath.Join(".claude", "commands", "btp-iac", "govern.md"),
		filepath.Join(".claude", "commands", "btp-iac", "generate.md"),
	}
	for _, rel := range expected {
		if _, err := os.Stat(filepath.Join("myproject", rel)); err != nil {
			t.Errorf("missing: myproject/%s", rel)
		}
	}
	// Cursor dir must NOT be created.
	if _, err := os.Stat(filepath.Join("myproject", ".cursor")); err == nil {
		t.Error("unexpected .cursor dir created for claude-only selection")
	}
}

func TestScaffoldCursor(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig) //nolint:errcheck

	agents := []Agent{KnownAgents["cursor"]}
	if err := Scaffold("myproject", assets.Commands, agents); err != nil {
		t.Fatalf("Scaffold: %v", err)
	}

	if _, err := os.Stat(filepath.Join("myproject", ".cursor", "rules", "govern.mdc")); err != nil {
		t.Errorf("missing cursor rule file: %v", err)
	}
	if _, err := os.Stat(filepath.Join("myproject", ".claude")); err == nil {
		t.Error("unexpected .claude dir created for cursor-only selection")
	}
}

func TestScaffoldCopilot(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig) //nolint:errcheck

	agents := []Agent{KnownAgents["copilot"]}
	if err := Scaffold("myproject", assets.Commands, agents); err != nil {
		t.Fatalf("Scaffold: %v", err)
	}

	if _, err := os.Stat(filepath.Join("myproject", ".github", "instructions", "govern.instructions.md")); err != nil {
		t.Errorf("missing copilot instructions file: %v", err)
	}
}

func TestScaffoldMultiAgent(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig) //nolint:errcheck

	agents := []Agent{KnownAgents["claude"], KnownAgents["cursor"], KnownAgents["copilot"]}
	if err := Scaffold("myproject", assets.Commands, agents); err != nil {
		t.Fatalf("Scaffold: %v", err)
	}

	checks := []string{
		filepath.Join(".claude", "commands", "btp-iac", "scenario.md"),
		filepath.Join(".cursor", "rules", "scenario.mdc"),
		filepath.Join(".github", "instructions", "scenario.instructions.md"),
	}
	for _, rel := range checks {
		if _, err := os.Stat(filepath.Join("myproject", rel)); err != nil {
			t.Errorf("missing: myproject/%s", rel)
		}
	}
}

func TestScaffoldRefusesExistingDir(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig) //nolint:errcheck

	if err := os.Mkdir("exists", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Scaffold("exists", assets.Commands, []Agent{KnownAgents["claude"]}); err == nil {
		t.Error("expected error when directory already exists, got nil")
	}
}
