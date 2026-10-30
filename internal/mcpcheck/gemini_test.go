package mcpcheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeminiMCPDiscovery(t *testing.T) {
	for _, tc := range []struct {
		name, user, project    string
		wantTerraform, wantBTP bool
	}{
		{"user", "Terraform-MCP,BTP-admin", "", true, true},
		{"project", "", "OpenTofu-server,BTP-admin", true, true},
		{"combined", "Terraform-MCP", "BTP-admin", true, true},
		{"missing", "", "", false, false},
		{"unrelated", "fetch", "sap-docs", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home, parent := t.TempDir(), t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			t.Chdir(parent)
			target := filepath.Join(parent, "project")
			if tc.user != "" {
				writeMCPConfig(t, home, ".gemini/settings.json", "mcpServers", strings.Split(tc.user, ","))
			}
			if tc.project != "" {
				writeMCPConfig(t, target, ".gemini/settings.json", "mcpServers", strings.Split(tc.project, ","))
			}
			// A server in the invoking directory must not satisfy the target.
			writeMCPConfig(t, parent, ".gemini/settings.json", "mcpServers", []string{"terraform-parent", "btp-parent"})
			warnings := Run(NonBTPChecks, []string{"gemini"}, target)
			if (len(warnings) == 0) != tc.wantTerraform {
				t.Fatalf("Terraform discovery: %v", warnings)
			}
			if BTPMCPAvailable("gemini", target) != tc.wantBTP {
				t.Fatal("unexpected BTP discovery")
			}
			if cwd, err := os.Getwd(); err != nil || cwd != parent {
				t.Fatalf("discovery changed working directory: %s, %v", cwd, err)
			}
		})
	}
}

func TestGeminiInvalidConfigDoesNotHideOtherScope(t *testing.T) {
	for _, invalid := range []string{"user", "project"} {
		t.Run(invalid, func(t *testing.T) {
			home, target := t.TempDir(), t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			userPath := writeMCPConfig(t, home, ".gemini/settings.json", "mcpServers", []string{"terraform", "btp"})
			projectPath := writeMCPConfig(t, target, ".gemini/settings.json", "mcpServers", []string{"opentofu", "btp"})
			invalidPath := userPath
			if invalid == "project" {
				invalidPath = projectPath
			}
			if err := os.WriteFile(invalidPath, []byte("invalid JSON"), 0o644); err != nil {
				t.Fatal(err)
			}
			if !BTPMCPAvailable("gemini", target) || len(Run(NonBTPChecks, []string{"gemini"}, target)) != 0 {
				t.Fatal("invalid configuration hid valid other-scope servers")
			}
		})
	}
}

func TestGeminiUnreadableConfig(t *testing.T) {
	home, target := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	// A directory cannot be read as a JSON settings file, on all test platforms.
	if err := os.MkdirAll(filepath.Join(target, ".gemini", "settings.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	if BTPMCPAvailable("gemini", target) {
		t.Fatal("unreadable configuration supplied a BTP route")
	}
	if warnings := Run(NonBTPChecks, []string{"gemini"}, target); len(warnings) != 1 || !strings.HasPrefix(warnings[0], "[gemini]") {
		t.Fatalf("unexpected advisory warnings: %v", warnings)
	}
}

func TestGeminiEmptyProjectDirUsesCurrentDirectory(t *testing.T) {
	home, project := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Chdir(project)
	writeMCPConfig(t, project, ".gemini/settings.json", "mcpServers", []string{"terraform", "btp"})
	if !BTPMCPAvailable("gemini", "") || len(Run(NonBTPChecks, []string{"gemini"}, "")) != 0 {
		t.Fatal("empty project directory did not discover current-directory servers")
	}
}
