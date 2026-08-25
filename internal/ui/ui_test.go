package ui

import (
	"strings"
	"testing"
)

func TestPrintSuccessContainsCommands(t *testing.T) {
	var buf strings.Builder
	printSuccess(&buf, "my-project", []string{"claude", "cursor"})
	output := buf.String()

	for _, cmd := range []string{
		"btp-iac.scenario", "btp-iac.accounts", "btp-iac.services",
		"btp-iac.security", "btp-iac.tasks", "btp-iac.design",
		"btp-iac.generate", "btp-iac.govern", "btp-iac.analyse",
	} {
		if !strings.Contains(output, cmd) {
			t.Errorf("output missing command %q", cmd)
		}
	}
}

func TestPrintSuccessContainsName(t *testing.T) {
	var buf strings.Builder
	printSuccess(&buf, "my-project", []string{"claude"})
	output := buf.String()

	if !strings.Contains(output, "my-project") {
		t.Error("output missing project name")
	}
	if !strings.Contains(output, "claude") {
		t.Error("output missing agent ID")
	}
}
