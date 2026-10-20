package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SAP/btp-iac-spec-kit/internal/scaffold"
)

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
