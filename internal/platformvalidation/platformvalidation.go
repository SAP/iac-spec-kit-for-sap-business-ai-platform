// Package platformvalidation records optional BTP platform validation routes
// selected during sap-iac initialization.
package platformvalidation

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SAP/btp-iac-spec-kit/internal/mcpcheck"
	"github.com/SAP/btp-iac-spec-kit/internal/preflight"
)

const (
	Directory = ".sap-iac"
	Filename  = "platform-validation.md"
)

// Ignore is the Git ignore pattern for the local capability record.
var Ignore = filepath.ToSlash(filepath.Join(Directory, Filename))

// Capabilities is the non-sensitive result of initialization pre-flight.
type Capabilities struct {
	CLIAvailable bool
	MCPAgents    map[string]bool
}

// Detect determines the optional BTP validation routes available to agents.
func Detect(agentIDs []string) Capabilities {
	c := Capabilities{CLIAvailable: preflight.Available("btp"), MCPAgents: make(map[string]bool, len(agentIDs))}
	for _, agent := range agentIDs {
		c.MCPAgents[agent] = mcpcheck.BTPMCPAvailable(agent)
	}
	return c
}

// HasRoute reports whether this agent can validate against BTP. The CLI is
// shared by all agents and is always preferred when it was found at init time.
func (c Capabilities) HasRoute(agent string) bool {
	return c.CLIAvailable || c.MCPAgents[agent]
}

// Write persists the detected capability record inside the initialized project.
func (c Capabilities) Write(projectDir string) error {
	dir := filepath.Join(projectDir, Directory)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create platform validation metadata: %w", err)
	}
	var mcpAgents []string
	for agent, available := range c.MCPAgents {
		if available {
			mcpAgents = append(mcpAgents, agent)
		}
	}
	sort.Strings(mcpAgents)
	cli := "unavailable"
	preferred := "none — accept user input without live platform validation"
	if c.CLIAvailable {
		cli = "available"
		preferred = "BTP CLI"
	} else if len(mcpAgents) > 0 {
		preferred = "BTP MCP for configured agents"
	}
	mcp := "none"
	if len(mcpAgents) > 0 {
		mcp = strings.Join(mcpAgents, ", ")
	}
	content := fmt.Sprintf(`# BTP Platform Validation

This local file records optional validation capabilities detected by sap-iac init. It contains no credentials.

- BTP CLI: %s
- BTP MCP agents: %s
- Preferred route: %s
`, cli, mcp, preferred)
	if err := os.WriteFile(filepath.Join(dir, Filename), []byte(content), 0o644); err != nil {
		return fmt.Errorf("write platform validation metadata: %w", err)
	}
	return nil
}

// Warnings returns a per-agent message only when neither optional route exists.
func (c Capabilities) Warnings(agentIDs []string) []string {
	var warnings []string
	for _, agent := range agentIDs {
		if !c.HasRoute(agent) {
			warnings = append(warnings, fmt.Sprintf("[%s] No BTP CLI or BTP MCP server detected. Platform validation will use user input without blocking.", agent))
		}
	}
	return warnings
}
