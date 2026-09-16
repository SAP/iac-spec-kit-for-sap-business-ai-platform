// Package agentselect resolves the set of AI agents to configure during init,
// either from the --agent flag or an interactive multi-select prompt.
package agentselect

import (
	"fmt"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/AlecAivazis/survey/v2"
	"github.com/SAP/btp-iac-spec-kit/internal/scaffold"
	"golang.org/x/term"
)

var promptStyle = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(12)).Bold(true)

var agentOrder = []string{"claude", "codex", "cursor", "copilot"}

// CheckTTYOrFlag returns an error if flagValue is empty and stdin is not a TTY.
// Call this before showing any interactive prompts to fail fast in CI/non-TTY
// environments.
func CheckTTYOrFlag(flagValue string) error {
	if flagValue != "" {
		return nil
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("no TTY detected: use --agent to specify agents (e.g. --agent claude,codex,cursor,copilot)")
	}
	return nil
}

// Select resolves the agents to install.
//
// If flagValue is non-empty it is parsed as a comma-separated list of agent IDs.
// If flagValue is empty and stdin is a TTY, an interactive prompt is shown.
// If flagValue is empty and stdin is not a TTY, an error is returned directing
// the caller to use --agent.
func Select(flagValue string) ([]scaffold.Agent, error) {
	if flagValue != "" {
		return fromFlag(flagValue)
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return nil, fmt.Errorf("no TTY detected: use --agent to specify agents (e.g. --agent claude,codex,cursor,copilot)")
	}
	return fromPrompt()
}

func fromFlag(flagValue string) ([]scaffold.Agent, error) {
	if strings.TrimSpace(flagValue) == "" {
		return nil, fmt.Errorf("--agent value is empty — supported: %s", strings.Join(agentOrder, ", "))
	}
	ids := strings.Split(flagValue, ",")
	agents := make([]scaffold.Agent, 0, len(ids))
	for _, raw := range ids {
		id := strings.TrimSpace(raw)
		a, ok := scaffold.KnownAgents[id]
		if !ok {
			return nil, fmt.Errorf("unknown agent %q — supported: %s", id, strings.Join(agentOrder, ", "))
		}
		agents = append(agents, a)
	}
	return agents, nil
}

func fromPrompt() ([]scaffold.Agent, error) {
	var selected []string
	prompt := &survey.MultiSelect{
		Message: promptStyle.Render("Select AI agents to configure:"),
		Options: agentOrder,
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		return nil, fmt.Errorf("prompt: %w", err)
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("at least one agent must be selected")
	}
	agents := make([]scaffold.Agent, 0, len(selected))
	for _, id := range selected {
		agents = append(agents, scaffold.KnownAgents[id])
	}
	return agents, nil
}
