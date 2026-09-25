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
		"sap-iac.scenario", "sap-iac.accounts", "sap-iac.services",
		"sap-iac.security", "sap-iac.connectivity", "sap-iac.tasks",
		"sap-iac.design", "sap-iac.generate", "sap-iac.govern", "sap-iac.analyse",
		"sap-iac.next",
	} {
		if !strings.Contains(output, cmd) {
			t.Errorf("output missing command %q", cmd)
		}
	}
}

func TestPrintSuccessAlignsUtilityCommand(t *testing.T) {
	var buf strings.Builder
	printSuccess(&buf, "my-project", []string{"claude"}, scaffold.ModeFresh, true)

	commandColumn := func(command string) int {
		for line := range strings.SplitSeq(buf.String(), "\n") {
			if column := strings.Index(line, command); column >= 0 {
				return column
			}
		}
		t.Fatalf("output missing command %q", command)
		return 0
	}

	if got, want := commandColumn("sap-iac.next"), commandColumn("sap-iac.scenario"); got != want {
		t.Errorf("sap-iac.next begins in column %d, want %d", got, want)
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
