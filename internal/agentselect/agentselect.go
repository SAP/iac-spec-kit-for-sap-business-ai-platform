// Package agentselect resolves the set of AI agents to configure during init,
// either from the --agent flag or an interactive multi-select prompt.
package agentselect

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/SAP/btp-iac-sdd/internal/scaffold"
	"golang.org/x/term"
)

var agentOrder = []string{"claude", "cursor", "copilot"}

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
		return nil, fmt.Errorf("no TTY detected: use --agent to specify agents (e.g. --agent claude,cursor,copilot)")
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
		Message: "Select AI agents to configure:",
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
