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
