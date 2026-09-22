package ui

import (
	"strings"
	"testing"

	"github.com/SAP/btp-iac-spec-kit/internal/scaffold"
)

func TestPrintSuccessContainsCommands(t *testing.T) {
	var buf strings.Builder
	printSuccess(&buf, "my-project", []string{"claude", "cursor"}, scaffold.ModeFresh, true)
	output := buf.String()

	for _, cmd := range []string{
		"btp-iac.scenario", "btp-iac.accounts", "btp-iac.services",
		"btp-iac.security", "btp-iac.connectivity", "btp-iac.tasks",
		"btp-iac.design", "btp-iac.generate", "btp-iac.govern", "btp-iac.analyse",
	} {
		if !strings.Contains(output, cmd) {
			t.Errorf("output missing command %q", cmd)
		}
	}
}

func TestPrintSuccessContainsName(t *testing.T) {
	var buf strings.Builder
	printSuccess(&buf, "my-project", []string{"claude"}, scaffold.ModeFresh, true)
	output := buf.String()

	if !strings.Contains(output, "my-project") {
		t.Error("output missing project name")
	}
	if !strings.Contains(output, "claude") {
		t.Error("output missing agent ID")
	}
}

func TestPrintSuccessUpdated(t *testing.T) {
	var buf strings.Builder
	printSuccess(&buf, "", []string{"cursor"}, scaffold.ModeAgentOnly, false)
	output := buf.String()

	if !strings.Contains(output, "Project updated") {
		t.Error("AgentOnly: expected 'Project updated'")
	}
	if strings.Contains(output, "cd ") {
		t.Error("AgentOnly: unexpected cd step in output")
	}
}
