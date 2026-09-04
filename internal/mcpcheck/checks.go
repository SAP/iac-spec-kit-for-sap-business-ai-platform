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
		return hasMatch(serverNamesForAgent(agent), "btp")
	},
	Warning: `No BTP Administration MCP server detected.
BTP account operations in your AI agent may not work.`,
}
