// Package backendselect resolves the Terraform backend type during init,
// either from the --backend flag or an interactive select prompt.
package backendselect

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/AlecAivazis/survey/v2"
	"github.com/SAP/btp-iac-spec-kit/internal/scaffold"
)

var promptStyle = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(12)).Bold(true)

const skipOption = "skip (no backend.tf)"

// Select resolves the backend type to use.
//
// If flagValue is non-empty it is validated and returned.
// If flagValue is empty an interactive prompt is shown; the user may skip.
// Returns an empty string when no backend was selected.
func Select(flagValue string) (string, error) {
	if flagValue != "" {
		return fromFlag(flagValue)
	}
	return fromPrompt()
}

func fromFlag(flagValue string) (string, error) {
	bt := strings.TrimSpace(flagValue)
	for _, known := range scaffold.KnownBackends {
		if bt == known {
			return bt, nil
		}
	}
	return "", fmt.Errorf("unknown backend type %q — supported: %s", bt, strings.Join(scaffold.KnownBackends, ", "))
}

func fromPrompt() (string, error) {
	options := append([]string{skipOption}, scaffold.KnownBackends...)
	var choice string
	if err := survey.AskOne(&survey.Select{
		Message: promptStyle.Render("Terraform backend type (for remote state):"),
		Options: options,
		Default: skipOption,
	}, &choice); err != nil {
		return "", fmt.Errorf("prompt: %w", err)
	}
	if choice == skipOption {
		return "", nil
	}
	return choice, nil
}
