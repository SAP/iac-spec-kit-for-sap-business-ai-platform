package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SAP/btp-iac-spec-kit/skills"
)

func TestScaffoldClaude(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig) //nolint:errcheck

	agents := []Agent{KnownAgents["claude"]}
	if _, err := Scaffold("myproject", skills.Commands, agents); err != nil {
		t.Fatalf("Scaffold: %v", err)
	}

	expected := []string{
		"specs", "memory", "terraform",
		filepath.Join(".claude", "commands"),
		".gitignore", ".git",
		filepath.Join(".claude", "commands", "btp-iac.govern.md"),
		filepath.Join(".claude", "commands", "btp-iac.generate.md"),
	}
	for _, rel := range expected {
		if _, err := os.Stat(filepath.Join("myproject", rel)); err != nil {
			t.Errorf("missing: myproject/%s", rel)
		}
	}

	// All ten command files must be present.
	allCmds := []string{
		"btp-iac.govern.md", "btp-iac.scenario.md", "btp-iac.analyse.md",
		"btp-iac.accounts.md", "btp-iac.services.md", "btp-iac.security.md",
		"btp-iac.connectivity.md", "btp-iac.tasks.md", "btp-iac.design.md",
		"btp-iac.generate.md",
	}
	for _, f := range allCmds {
		if _, err := os.Stat(filepath.Join("myproject", ".claude", "commands", f)); err != nil {
			t.Errorf("missing command file: .claude/commands/%s", f)
		}
	}
	entries, err := os.ReadDir(filepath.Join("myproject", ".claude", "commands"))
	if err != nil {
		t.Fatalf("ReadDir .claude/commands: %v", err)
	}
	if got := len(entries); got != len(allCmds) {
		t.Errorf(".claude/commands has %d files, want %d", got, len(allCmds))
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
	if _, err := Scaffold("myproject", skills.Commands, agents); err != nil {
		t.Fatalf("Scaffold: %v", err)
	}

	if _, err := os.Stat(filepath.Join("myproject", ".cursor", "rules", "btp-iac.govern.mdc")); err != nil {
		t.Errorf("missing cursor rule file: %v", err)
	}
	if _, err := os.Stat(filepath.Join("myproject", ".claude")); err == nil {
		t.Error("unexpected .claude dir created for cursor-only selection")
	}
}

func TestScaffoldCodex(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig) //nolint:errcheck

	agents := []Agent{KnownAgents["codex"]}
	if _, err := Scaffold("myproject", skills.Commands, agents); err != nil {
		t.Fatalf("Scaffold: %v", err)
	}

	if _, err := os.Stat(filepath.Join("myproject", ".codex", "prompts", "btp-iac.govern.md")); err != nil {
		t.Errorf("missing codex prompt file: %v", err)
	}
	if _, err := os.Stat(filepath.Join("myproject", ".claude")); err == nil {
		t.Error("unexpected .claude dir created for codex-only selection")
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
	if _, err := Scaffold("myproject", skills.Commands, agents); err != nil {
		t.Fatalf("Scaffold: %v", err)
	}

	if _, err := os.Stat(filepath.Join("myproject", ".github", "instructions", "btp-iac.govern.instructions.md")); err != nil {
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

	agents := []Agent{KnownAgents["claude"], KnownAgents["codex"], KnownAgents["cursor"], KnownAgents["copilot"]}
	if _, err := Scaffold("myproject", skills.Commands, agents); err != nil {
		t.Fatalf("Scaffold: %v", err)
	}

	checks := []string{
		filepath.Join(".claude", "commands", "btp-iac.scenario.md"),
		filepath.Join(".codex", "prompts", "btp-iac.scenario.md"),
		filepath.Join(".cursor", "rules", "btp-iac.scenario.mdc"),
		filepath.Join(".github", "instructions", "btp-iac.scenario.instructions.md"),
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
	if _, err := Scaffold("exists", skills.Commands, []Agent{KnownAgents["claude"]}); err == nil {
		t.Error("expected error when directory already exists, got nil")
	}
}

func TestScaffoldGitAbsent(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig) //nolint:errcheck

	// Hide git from PATH so LookPath fails.
	t.Setenv("PATH", "")

	agents := []Agent{KnownAgents["claude"]}
	notice, err := Scaffold("myproject", skills.Commands, agents)
	if err != nil {
		t.Fatalf("Scaffold should succeed without git, got: %v", err)
	}
	if notice == "" {
		t.Error("expected non-empty notice when git is absent")
	}
	// Files created; no .git dir.
	if _, err := os.Stat(filepath.Join("myproject", ".gitignore")); err != nil {
		t.Error("missing .gitignore")
	}
	if _, err := os.Stat(filepath.Join("myproject", ".git")); err == nil {
		t.Error(".git dir should not exist when git was absent")
	}
}

func TestIsProject(t *testing.T) {
	dir := t.TempDir()

	if IsProject(dir) {
		t.Error("empty dir should not be detected as a project")
	}

	if err := os.Mkdir(filepath.Join(dir, "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !IsProject(dir) {
		t.Error("dir with specs/ should be detected as a project")
	}
}

func TestApplyAdopt(t *testing.T) {
	dir := t.TempDir()
	// Pre-existing .tf file at root — simulates an existing Terraform repo.
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte("# existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	agents := []Agent{KnownAgents["claude"]}
	if _, err := Apply(dir, ModeAdopt, skills.Commands, agents); err != nil {
		t.Fatalf("Apply(Adopt): %v", err)
	}

	mustExist := []string{
		"specs", "memory", "terraform",
		filepath.Join(".claude", "commands", "btp-iac.govern.md"),
		".gitignore",
	}
	for _, rel := range mustExist {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Errorf("Adopt: missing %s", rel)
		}
	}
	// Pre-existing file must be untouched.
	if _, err := os.Stat(filepath.Join(dir, "main.tf")); err != nil {
		t.Error("Adopt: main.tf was removed")
	}
	// git init must NOT run (no .git created).
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		t.Error("Adopt: unexpected .git created")
	}
}

func TestApplyAdoptPreservesGitignore(t *testing.T) {
	dir := t.TempDir()
	existing := "# custom\n"
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Apply(dir, ModeAdopt, skills.Commands, []Agent{KnownAgents["claude"]}); err != nil {
		t.Fatalf("Apply(Adopt): %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	want := existing + ".btp-iac/platform-validation.md\nmemory/global-account.md\n"
	if string(got) != want {
		t.Errorf("Adopt: .gitignore = %q, want %q", string(got), want)
	}
}

func TestEnsureGitignoreEntryRecognizesCRLFAndDirectoryRule(t *testing.T) {
	tests := []struct {
		existing string
		entry    string
	}{
		{existing: ".btp-iac/platform-validation.md\r\n", entry: ".btp-iac/platform-validation.md"},
		{existing: ".btp-iac/\n", entry: ".btp-iac/platform-validation.md"},
		{existing: "memory/\n", entry: filepath.Join("memory", GlobalAccountFile)},
	}
	for _, test := range tests {
		dir := t.TempDir()
		path := filepath.Join(dir, ".gitignore")
		if err := os.WriteFile(path, []byte(test.existing), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := ensureGitignoreEntry(dir, test.entry); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != test.existing {
			t.Errorf(".gitignore changed from %q to %q", test.existing, got)
		}
	}
}

func TestApplyAgentOnly(t *testing.T) {
	dir := t.TempDir()
	// Simulate an existing btp-iac project — specs and memory already exist.
	for _, d := range []string{"specs", "memory", "terraform"} {
		if err := os.Mkdir(filepath.Join(dir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	// Add cursor on top of an existing project.
	if _, err := Apply(dir, ModeAgentOnly, skills.Commands, []Agent{KnownAgents["cursor"]}); err != nil {
		t.Fatalf("Apply(AgentOnly): %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".cursor", "rules", "btp-iac.govern.mdc")); err != nil {
		t.Error("AgentOnly: missing cursor rule file")
	}
	// AgentOnly records local files that must remain untracked too.
	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil || !strings.Contains(string(data), ".btp-iac/platform-validation.md") || !strings.Contains(string(data), "memory/global-account.md") {
		t.Errorf("AgentOnly: local-file ignore rules missing: %v", err)
	}
}

func TestWriteGlobalAccountSubdomain(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "memory"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := WriteGlobalAccountSubdomain(dir, "acme-global"); err != nil {
		t.Fatalf("WriteGlobalAccountSubdomain: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "memory", GlobalAccountFile))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), "# Global Account\n\n- Subdomain: acme-global\n"; got != want {
		t.Errorf("global account record = %q, want %q", got, want)
	}

	if err := WriteGlobalAccountSubdomain(dir, ""); err != nil {
		t.Fatalf("WriteGlobalAccountSubdomain(empty): %v", err)
	}
	data, err = os.ReadFile(filepath.Join(dir, "memory", GlobalAccountFile))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "acme-global") {
		t.Error("empty optional input overwrote the configured subdomain")
	}

	emptyDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(emptyDir, "memory"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := WriteGlobalAccountSubdomain(emptyDir, ""); err != nil {
		t.Fatalf("WriteGlobalAccountSubdomain(empty new project): %v", err)
	}
	data, err = os.ReadFile(filepath.Join(emptyDir, "memory", GlobalAccountFile))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), "# Global Account\n\n- Subdomain:\n"; got != want {
		t.Errorf("empty global account record = %q, want %q", got, want)
	}
}
