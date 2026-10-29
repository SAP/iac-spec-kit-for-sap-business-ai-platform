package platformvalidation

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGeminiDetectedRoutesAndPrivateRecord(t *testing.T) {
	for _, cli := range []bool{false, true} {
		t.Run(map[bool]string{false: "MCP", true: "CLI preferred"}[cli], func(t *testing.T) {
			home, project, bin := t.TempDir(), t.TempDir(), t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			t.Setenv("PATH", bin)
			if cli {
				name := "btp"
				if runtime.GOOS == "windows" {
					name += ".exe"
				}
				if err := os.WriteFile(filepath.Join(bin, name), []byte("placeholder"), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			settingsDir := filepath.Join(project, ".gemini")
			if err := os.MkdirAll(settingsDir, 0o755); err != nil {
				t.Fatal(err)
			}
			settings := []byte(`{"mcpServers":{"BTP-admin":{"httpUrl":"https://private.example","env":{"TOKEN":"secret-token"}}}}`)
			if err := os.WriteFile(filepath.Join(settingsDir, "settings.json"), settings, 0o644); err != nil {
				t.Fatal(err)
			}
			capabilities := Detect([]string{"gemini"}, project)
			if !capabilities.MCPAgents["gemini"] || capabilities.CLIAvailable != cli || len(capabilities.Warnings([]string{"gemini"})) != 0 {
				t.Fatalf("unexpected capabilities: %+v", capabilities)
			}
			if err := capabilities.Write(project); err != nil {
				t.Fatal(err)
			}
			record, err := os.ReadFile(filepath.Join(project, Directory, Filename))
			if err != nil {
				t.Fatal(err)
			}
			preferred := "Preferred route: BTP MCP for configured agents"
			if cli {
				preferred = "Preferred route: BTP CLI"
			}
			if !strings.Contains(string(record), "BTP MCP agents: gemini") || !strings.Contains(string(record), preferred) {
				t.Fatalf("missing Gemini route: %s", record)
			}
			for _, private := range []string{"private.example", "TOKEN", "secret-token", "httpUrl"} {
				if strings.Contains(string(record), private) {
					t.Errorf("configuration leaked into record: %s", private)
				}
			}
		})
	}
}
