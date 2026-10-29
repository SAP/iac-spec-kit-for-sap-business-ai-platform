package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SAP/btp-iac-spec-kit/internal/scaffold"
)

func TestGeminiInitHelp(t *testing.T) {
	cmd := rootCmd()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"init", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "claude, codex, cursor, copilot, gemini") {
		t.Fatalf("help missing supported agent set: %s", &output)
	}
}

func TestGeminiInitNamedProjectRefresh(t *testing.T) {
	parent, home := t.TempDir(), t.TempDir()
	t.Chdir(parent)
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	project := filepath.Join(parent, "project")
	for _, rel := range []string{"specs/scenario.md", "memory/global-account.md", "terraform/main.tf", ".gemini/commands/custom.toml"} {
		path := filepath.Join(project, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("user content"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	settings := []byte(`{"mcpServers":{"Terraform-MCP":{"command":"fake"},"BTP-MCP":{"command":"fake","env":{"TOKEN":"private-token"}}}}`)
	settingsPath := filepath.Join(project, ".gemini", "settings.json")
	if err := os.WriteFile(settingsPath, settings, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, configured := range []bool{true, false} {
		if !configured {
			if err := os.WriteFile(settingsPath, []byte("{}"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		cmd := rootCmd()
		var warnings bytes.Buffer
		cmd.SetErr(&warnings)
		cmd.SetArgs([]string{"init", "project", "--agent", "gemini"})
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(warnings.String(), "No Terraform or OpenTofu MCP") == configured {
			t.Fatalf("wrong target-project warning, configured=%v: %s", configured, &warnings)
		}
		for _, rel := range []string{".gemini/commands/sap-iac.scenario.toml", ".gemini/skills/sap-iac-scenario/SKILL.md"} {
			if _, err := os.Stat(filepath.Join(project, filepath.FromSlash(rel))); err != nil {
				t.Fatal(err)
			}
		}
		record, err := os.ReadFile(filepath.Join(project, ".sap-iac", "platform-validation.md"))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(record), "BTP MCP agents: gemini") != configured || strings.Contains(string(record), "private-token") {
			t.Fatalf("incorrect refreshed record: %s", record)
		}
	}
	for _, rel := range []string{"specs/scenario.md", "memory/global-account.md", "terraform/main.tf", ".gemini/commands/custom.toml"} {
		data, err := os.ReadFile(filepath.Join(project, filepath.FromSlash(rel)))
		if err != nil || string(data) != "user content" {
			t.Fatalf("refresh changed %s", rel)
		}
	}
}

func TestExistingDirIntent(t *testing.T) {
	dir := t.TempDir()
	if got := existingDirIntent(dir); got.mode != scaffold.ModeAdopt {
		t.Errorf("bare directory mode = %v, want ModeAdopt", got.mode)
	}

	if err := os.Mkdir(filepath.Join(dir, "memory"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := existingDirIntent(dir); got.mode != scaffold.ModeAgentOnly {
		t.Errorf("existing project mode = %v, want ModeAgentOnly", got.mode)
	}
}

func TestProjectNameFromURL(t *testing.T) {
	cases := []struct {
		url     string
		want    string
		wantErr bool
	}{
		{"https://github.com/org/my-project.git", "my-project", false},
		{"https://github.com/org/my-project", "my-project", false},
		{"git@github.com:org/my-project.git", "my-project", false},
		{"git@github.com:org/my-project", "my-project", false},
		{"https://github.com/org/repo/", "repo", false},
		{"https://github.com/org/repo.git/", "repo", false},
		{"/", "", true},
	}
	for _, tc := range cases {
		got, err := projectNameFromURL(tc.url)
		if (err != nil) != tc.wantErr {
			t.Errorf("projectNameFromURL(%q) error = %v, wantErr %v", tc.url, err, tc.wantErr)
			continue
		}
		if got != tc.want {
			t.Errorf("projectNameFromURL(%q) = %q, want %q", tc.url, got, tc.want)
		}
	}
}

func TestValidBranchName(t *testing.T) {
	for _, tc := range []struct {
		name    string
		wantErr bool
	}{
		{"main", false},
		{"feature/iac-setup", false},
		{"", true},
		{"feature..setup", true},
	} {
		if err := validBranchName(tc.name); (err != nil) != tc.wantErr {
			t.Errorf("validBranchName(%q) error = %v, wantErr %v", tc.name, err, tc.wantErr)
		}
	}
}

func TestAppendToCatalogue_AddsNewlineWhenMissing(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "catalogue.yaml")

	// Write content without a trailing newline.
	if err := os.WriteFile(dest, []byte("- service: foo"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := appendToCatalogue(dest, []byte("- service: bar\n")); err != nil {
		t.Fatalf("appendToCatalogue: %v", err)
	}
	got, _ := os.ReadFile(dest)
	want := "- service: foo\n- service: bar\n"
	if string(got) != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestAppendToCatalogue_NoExtraNewlineWhenPresent(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "catalogue.yaml")

	if err := os.WriteFile(dest, []byte("- service: foo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := appendToCatalogue(dest, []byte("- service: bar\n")); err != nil {
		t.Fatalf("appendToCatalogue: %v", err)
	}
	got, _ := os.ReadFile(dest)
	want := "- service: foo\n- service: bar\n"
	if string(got) != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestAppendToCatalogue_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "catalogue.yaml")

	if err := appendToCatalogue(dest, []byte("- service: bar\n")); err != nil {
		t.Fatalf("appendToCatalogue on empty file: %v", err)
	}
	got, _ := os.ReadFile(dest)
	if string(got) != "- service: bar\n" {
		t.Errorf("unexpected content: %s", got)
	}
}
