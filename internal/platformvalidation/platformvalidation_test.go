package platformvalidation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteRecordsCapabilities(t *testing.T) {
	dir := t.TempDir()
	c := Capabilities{CLIAvailable: true, MCPAgents: map[string]bool{"cursor": true, "claude": true, "codex": false}}
	if err := c.Write(dir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, Directory, Filename))
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	for _, want := range []string{"BTP CLI: available", "BTP MCP agents: claude, cursor", "Preferred route: BTP CLI"} {
		if !strings.Contains(got, want) {
			t.Errorf("record missing %q: %s", want, got)
		}
	}
}

func TestWarningsAndRoutes(t *testing.T) {
	c := Capabilities{MCPAgents: map[string]bool{"claude": true, "cursor": false}}
	if !c.HasRoute("claude") || c.HasRoute("cursor") {
		t.Fatal("unexpected route selection")
	}
	warnings := c.Warnings([]string{"claude", "cursor"})
	if len(warnings) != 1 || !strings.HasPrefix(warnings[0], "[cursor]") {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
}
