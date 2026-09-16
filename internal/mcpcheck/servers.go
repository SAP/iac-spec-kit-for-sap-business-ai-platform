package mcpcheck

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	_ "modernc.org/sqlite" // pure-Go SQLite driver, no cgo required
)

// serverKeysFromFile reads top-level server keys from a JSON config file.
// It tries "mcpServers" first (Claude Code / Cursor format), then "servers"
// (VS Code / Copilot format). Returns nil if the file is absent or unparseable.
func serverKeysFromFile(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var cfg struct {
		MCPServers map[string]json.RawMessage `json:"mcpServers"` // Claude Code, Cursor
		Servers    map[string]json.RawMessage `json:"servers"`    // VS Code, Copilot
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil
	}
	merged := make(map[string]struct{}, len(cfg.MCPServers)+len(cfg.Servers))
	for k := range cfg.MCPServers {
		merged[k] = struct{}{}
	}
	for k := range cfg.Servers {
		merged[k] = struct{}{}
	}
	keys := make([]string, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	return keys
}

// codexServerNames returns MCP server names from Codex config files:
//   - ~/.codex/config.toml      (user-level)
//   - .codex/config.toml in cwd (project-level)
func codexServerNames() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	var names []string
	names = append(names, codexServerNamesFromTOML(filepath.Join(home, ".codex", "config.toml"))...)
	names = append(names, codexServerNamesFromTOML(filepath.Join(".codex", "config.toml"))...)
	return names
}

// codexServerNamesFromTOML extracts top-level [mcp_servers.<name>] table names
// from a Codex config.toml. It ignores nested tables such as
// [mcp_servers.<name>.env].
func codexServerNamesFromTOML(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var names []string
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name, ok := codexMCPTableName(line)
		if ok {
			names = append(names, name)
		}
	}
	return names
}

func codexMCPTableName(line string) (string, bool) {
	const prefix = "[mcp_servers."
	if !strings.HasPrefix(line, prefix) || !strings.HasSuffix(line, "]") {
		return "", false
	}
	body := strings.TrimSuffix(strings.TrimPrefix(line, prefix), "]")
	if body == "" {
		return "", false
	}
	if strings.HasPrefix(body, `"`) {
		name, err := strconv.Unquote(body)
		if err != nil || name == "" {
			return "", false
		}
		return name, true
	}
	if strings.Contains(body, ".") {
		return "", false
	}
	return body, true
}

// gatewayEntry is the shape of a Docker MCP gateway mcpServers entry.
type gatewayEntry struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// dockerGatewayProfileFromEntries scans raw mcpServer config values for a
// "docker mcp gateway run" entry and extracts the --profile value if present.
// Returns ("", false) when no gateway entry exists, ("", true) when the gateway
// is present but no --profile is specified (registry.yaml mode), and
// (id, true) when --profile <id> is set.
func dockerGatewayProfileFromEntries(entries map[string]json.RawMessage) (profileID string, found bool) {
	for _, raw := range entries {
		var e gatewayEntry
		if err := json.Unmarshal(raw, &e); err != nil {
			continue
		}
		if e.Command != "docker" || len(e.Args) < 3 {
			continue
		}
		if e.Args[0] != "mcp" || e.Args[1] != "gateway" || e.Args[2] != "run" {
			continue
		}
		// Gateway entry found. Look for --profile <id> or --profile=<id>.
		for i, arg := range e.Args {
			if arg == "--profile" && i+1 < len(e.Args) {
				return e.Args[i+1], true
			}
			if profileID, ok := strings.CutPrefix(arg, "--profile="); ok && profileID != "" {
				return profileID, true
			}
		}
		return "", true // gateway present, no --profile
	}
	return "", false
}

// mcpDockerServerNames returns the names of MCP servers that are active in the
// Docker MCP Gateway for a specific client configuration.
//
// When profileID is non-empty, the gateway was started with --profile <id> and
// the enabled servers are read from the working_set table (SQLite) filtered to
// that profile ID.
//
// When profileID is empty (no --profile arg), the gateway uses registry.yaml as
// its server list; that file's top-level keys under "registry:" are returned.
//
// Returns nil if the relevant source is absent or unreadable.
func mcpDockerServerNames(profileID string) []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	mcpDir := filepath.Join(home, ".docker", "mcp")

	if profileID == "" {
		return registryYAMLServerNames(filepath.Join(mcpDir, "registry.yaml"))
	}
	return workingSetServerNames(filepath.Join(mcpDir, "mcp-toolkit.db"), profileID)
}

// registryYAMLServerNames reads the server names that are enabled in the Docker
// MCP Gateway when it runs without a --profile flag. The registry.yaml format is:
//
//	registry:
//	  server-name:
//	    ref: ""
func registryYAMLServerNames(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var (
		names      []string
		inRegistry bool
	)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "registry:" {
			inRegistry = true
			continue
		}
		if !inRegistry {
			continue
		}
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' && line[0] != '#' {
			break // new top-level key ends the registry block
		}
		// Server names sit at exactly two-space indent: "  name:"
		if strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "   ") {
			name, _, found := strings.Cut(strings.TrimPrefix(line, "  "), ":")
			if found && name != "" && !strings.HasPrefix(name, "#") {
				names = append(names, name)
			}
		}
	}
	return names
}

// workingSetServerNames queries the Docker MCP Gateway's SQLite database for
// the servers enabled in the named profile.
func workingSetServerNames(dbPath, profileID string) []string {
	if _, err := os.Stat(dbPath); err != nil {
		return nil
	}
	conn, err := sql.Open("sqlite", "file:"+dbPath+"?mode=ro")
	if err != nil {
		return nil
	}
	defer func() { _ = conn.Close() }()

	const query = `SELECT json_extract(value, '$.snapshot.server.name')
	               FROM working_set, json_each(working_set.servers)
	               WHERE working_set.id = ?
	                 AND json_extract(value, '$.snapshot.server.name') IS NOT NULL`
	rows, err := conn.Query(query, profileID)
	if err != nil {
		return nil
	}
	defer func() { _ = rows.Close() }()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil && name != "" {
			names = append(names, name)
		}
	}
	if rows.Err() != nil {
		return nil
	}
	return names
}

// claudeGatewayProfile returns the Docker MCP gateway profile configured for
// Claude Code, scanning both the user-level and project-level mcpServers entries
// in ~/.claude.json.
// Returns ("", false) when no Docker gateway entry is found.
func claudeGatewayProfile() (profileID string, found bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", false
	}
	data, err := os.ReadFile(filepath.Join(home, ".claude.json"))
	if err != nil {
		return "", false
	}
	var cfg struct {
		MCPServers map[string]json.RawMessage `json:"mcpServers"`
		Projects   map[string]struct {
			MCPServers map[string]json.RawMessage `json:"mcpServers"`
		} `json:"projects"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", false
	}

	if profileID, found = dockerGatewayProfileFromEntries(cfg.MCPServers); found {
		return profileID, true
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", false
	}
	if resolved, err := filepath.EvalSymlinks(cwd); err == nil {
		cwd = resolved
	}
	if proj, ok := cfg.Projects[cwd]; ok {
		if profileID, found = dockerGatewayProfileFromEntries(proj.MCPServers); found {
			return profileID, true
		}
	}
	return "", false
}

// cursorGatewayProfile returns the Docker MCP gateway profile configured for
// Cursor, scanning ~/.cursor/mcp.json and .cursor/mcp.json in cwd.
// Returns ("", false) when no Docker gateway entry is found.
func cursorGatewayProfile() (profileID string, found bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", false
	}
	for _, path := range []string{
		filepath.Join(home, ".cursor", "mcp.json"),
		filepath.Join(".cursor", "mcp.json"),
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var cfg struct {
			MCPServers map[string]json.RawMessage `json:"mcpServers"`
		}
		if err := json.Unmarshal(data, &cfg); err != nil {
			continue
		}
		if profileID, found = dockerGatewayProfileFromEntries(cfg.MCPServers); found {
			return profileID, true
		}
	}
	return "", false
}

// claudeServerNames returns MCP server names from Claude Code config files:
//   - ~/.claude.json           (user scope: top-level mcpServers key)
//   - ~/.claude.json projects  (project scope: projects[cwd].mcpServers, set by `claude mcp add`)
//   - .mcp.json in cwd         (project scope: top-level mcpServers key, committed by teams)
func claudeServerNames() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	var names []string
	names = append(names, serverKeysFromFile(filepath.Join(home, ".claude.json"))...)
	names = append(names, claudeProjectServerNames(filepath.Join(home, ".claude.json"))...)
	names = append(names, serverKeysFromFile(".mcp.json")...)
	return names
}

// claudeProjectServerNames reads the project-scoped MCP servers from ~/.claude.json.
// `claude mcp add` (without --scope user) stores servers under projects[cwd].mcpServers.
func claudeProjectServerNames(claudeJSON string) []string {
	data, err := os.ReadFile(claudeJSON)
	if err != nil {
		return nil
	}
	var cfg struct {
		Projects map[string]struct {
			MCPServers map[string]json.RawMessage `json:"mcpServers"`
		} `json:"projects"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}
	// filepath.EvalSymlinks resolves macOS /var → /private/var differences so
	// the key lookup matches what `claude mcp add` wrote.
	if resolved, err := filepath.EvalSymlinks(cwd); err == nil {
		cwd = resolved
	}
	proj, ok := cfg.Projects[cwd]
	if !ok {
		return nil
	}
	keys := make([]string, 0, len(proj.MCPServers))
	for k := range proj.MCPServers {
		keys = append(keys, k)
	}
	return keys
}

// cursorServerNames returns MCP server names from Cursor config files:
//   - ~/.cursor/mcp.json        (global, all projects)
//   - .cursor/mcp.json in cwd   (project-level, takes priority in Cursor but we merge)
func cursorServerNames() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	var names []string
	names = append(names, serverKeysFromFile(filepath.Join(home, ".cursor", "mcp.json"))...)
	names = append(names, serverKeysFromFile(filepath.Join(".cursor", "mcp.json"))...)
	return names
}

// copilotServerNames returns MCP server names from the VS Code config locations:
//   - .vscode/mcp.json in cwd  (project-level, key: "servers")
//   - VS Code user mcp.json    (user-level, key: "servers", platform-dependent path)
func copilotServerNames() []string {
	var names []string
	names = append(names, serverKeysFromFile(filepath.Join(".vscode", "mcp.json"))...)
	if path := vsCodeMCPPath(); path != "" {
		names = append(names, serverKeysFromFile(path)...)
	}
	return names
}

// vsCodeMCPPath returns the platform-specific path to VS Code's user-level mcp.json.
func vsCodeMCPPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Code", "User", "mcp.json")
	case "windows":
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return filepath.Join(appdata, "Code", "User", "mcp.json")
		}
	default: // Linux
		if cfgDir := os.Getenv("XDG_CONFIG_HOME"); cfgDir != "" {
			return filepath.Join(cfgDir, "Code", "User", "mcp.json")
		}
		return filepath.Join(home, ".config", "Code", "User", "mcp.json")
	}
	return ""
}

// serverNamesForAgent returns MCP server names configured for a single agent.
func serverNamesForAgent(agent string) []string {
	switch agent {
	case "claude":
		names := claudeServerNames()
		if profileID, found := claudeGatewayProfile(); found {
			names = append(names, mcpDockerServerNames(profileID)...)
		}
		return names
	case "codex":
		return codexServerNames()
	case "cursor":
		names := cursorServerNames()
		if profileID, found := cursorGatewayProfile(); found {
			names = append(names, mcpDockerServerNames(profileID)...)
		}
		return names
	case "copilot":
		return copilotServerNames()
	default:
		return nil
	}
}

// hasMatch returns true if any name contains any of the substrings (case-insensitive).
func hasMatch(names []string, substrings ...string) bool {
	for _, n := range names {
		lower := strings.ToLower(n)
		for _, sub := range substrings {
			if strings.Contains(lower, sub) {
				return true
			}
		}
	}
	return false
}
