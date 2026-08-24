package ui

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestPrintSuccessBorderWidth(t *testing.T) {
	var buf strings.Builder
	printSuccess(&buf, "my-project", []string{"claude"})
	output := buf.String()

	const want = 82 // │ + space + 78 content + space + │

	for line := range strings.SplitSeq(output, "\n") {
		r := []rune(line)
		if len(r) == 0 {
			continue
		}
		first, last := string(r[0]), string(r[len(r)-1])
		if first == "│" || first == "├" || first == "╭" || first == "╰" {
			got := utf8.RuneCountInString(line)
			if got != want {
				t.Errorf("line width %d, want %d: %q", got, want, line)
			}
			_ = last
		}
	}
}

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
