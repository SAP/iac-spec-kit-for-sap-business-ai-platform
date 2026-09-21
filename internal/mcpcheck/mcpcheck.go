// Package mcpcheck probes the local environment for MCP server presence
// and returns warnings for any that are missing.
//
// BTP availability is consumed by platformvalidation through BTPMCPAvailable.
// To add another independent init warning, append a Check literal to
// NonBTPChecks. Set Agents to nil to match all agents; otherwise list the
// agent IDs it applies to.
package mcpcheck

import "slices"

// Check is a named MCP-presence probe scoped to a subset of agents.
type Check struct {
	// Name is the human-readable label used in warning messages.
	Name string
	// Agents lists the agent IDs this check applies to (nil = all agents).
	Agents []string
	// Probe returns true if the MCP server (or an acceptable alternative) is
	// present for the given single agent. Run calls it once per selected agent.
	Probe func(agent string) bool
	// Warning is emitted when Probe returns false.
	// Run prepends "[agentname] " so the caller knows which agent is affected.
	Warning string
}

// NonBTPChecks are the default MCP checks that remain independent from the
// combined CLI-or-MCP BTP platform-validation capability.
var NonBTPChecks = []Check{terraformCheck}

// Run executes every check that applies to at least one of the selected agents.
// Probe is called once per (check, agent) pair; a warning is emitted for each
// agent that fails, prefixed with "[agentname] ".
func Run(checks []Check, selectedAgents []string) []string {
	var warnings []string
	for _, c := range checks {
		for _, agent := range selectedAgents {
			if !applies(c.Agents, agent) {
				continue
			}
			if !c.Probe(agent) {
				warnings = append(warnings, "["+agent+"] "+c.Warning)
			}
		}
	}
	return warnings
}

// applies returns true when checkAgents is nil (all agents) or contains agent.
func applies(checkAgents []string, agent string) bool {
	return checkAgents == nil || slices.Contains(checkAgents, agent)
}
