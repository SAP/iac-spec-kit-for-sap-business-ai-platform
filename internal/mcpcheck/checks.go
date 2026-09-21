package mcpcheck

var terraformCheck = Check{
	Name:   "Terraform / OpenTofu MCP",
	Agents: nil, // applies to all agents
	Probe: func(agent string) bool {
		return hasMatch(serverNamesForAgent(agent), "terraform", "opentofu")
	},
	Warning: `No Terraform or OpenTofu MCP server detected.
Terraform operations in your AI agent may not work.`,
}

var btpAdminCheck = Check{
	Name:   "BTP Administration MCP",
	Agents: nil, // applies to all agents
	Probe: func(agent string) bool {
		return BTPMCPAvailable(agent)
	},
	Warning: `No BTP Administration MCP server detected.
BTP account operations in your AI agent may not work.`,
}

// BTPMCPAvailable reports whether the selected agent has a BTP MCP server
// configured. The server name is intentionally matched broadly because MCP
// server names are user-configurable.
func BTPMCPAvailable(agent string) bool {
	return hasMatch(serverNamesForAgent(agent), "btp")
}
