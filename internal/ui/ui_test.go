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
		"btp-iac.next",
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
		for _, line := range strings.Split(buf.String(), "\n") {
			if column := strings.Index(line, command); column >= 0 {
				return column
			}
		}
		t.Fatalf("output missing command %q", command)
		return 0
	}

	if got, want := commandColumn("btp-iac.next"), commandColumn("btp-iac.scenario"); got != want {
		t.Errorf("btp-iac.next begins in column %d, want %d", got, want)
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
