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
