package mcpcheck

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// writeMCPConfig writes a JSON MCP config file with the given key ("mcpServers"
// or "servers") and server names, returning the file path.
func writeMCPConfig(t *testing.T, dir, filename, key string, servers []string) string {
	t.Helper()
	m := make(map[string]any, len(servers))
	for _, s := range servers {
		m[s] = map[string]string{"command": "fake"}
	}
	data, err := json.Marshal(map[string]any{key: m})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, filename)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestApplies(t *testing.T) {
	tests := []struct {
		checkAgents []string
		agent       string
		want        bool
	}{
		{nil, "claude", true}, // nil = all agents
		{nil, "codex", true},
		{[]string{"claude"}, "claude", true},
		{[]string{"claude"}, "cursor", false},
		{[]string{"claude", "cursor"}, "copilot", false},
		{[]string{"claude", "cursor"}, "cursor", true},
	}
	for _, tt := range tests {
		if got := applies(tt.checkAgents, tt.agent); got != tt.want {
			t.Errorf("applies(%v, %q) = %v, want %v", tt.checkAgents, tt.agent, got, tt.want)
		}
	}
}

func TestRun_allPresent(t *testing.T) {
	checks := []Check{
		{Name: "A", Agents: nil, Probe: func(string) bool { return true }, Warning: "warn-a"},
		{Name: "B", Agents: nil, Probe: func(string) bool { return true }, Warning: "warn-b"},
	}
	if got := Run(checks, []string{"claude"}); len(got) != 0 {
		t.Errorf("expected no warnings, got %v", got)
	}
}

func TestRun_missingServer(t *testing.T) {
	checks := []Check{
		{Name: "A", Agents: nil, Probe: func(string) bool { return false }, Warning: "warn-a"},
		{Name: "B", Agents: nil, Probe: func(string) bool { return true }, Warning: "warn-b"},
	}
	got := Run(checks, []string{"claude"})
	if len(got) != 1 || got[0] != "[claude] warn-a" {
		t.Errorf("expected [[claude] warn-a], got %v", got)
	}
}

func TestRun_agentFilter(t *testing.T) {
	checks := []Check{
		{Name: "cursor-only", Agents: []string{"cursor"}, Probe: func(string) bool { return false }, Warning: "cursor-warn"},
	}
	// claude selected → check doesn't apply → no warning
	if got := Run(checks, []string{"claude"}); len(got) != 0 {
		t.Errorf("expected no warnings for non-matching agent, got %v", got)
	}
	// cursor selected → check applies → warning emitted with agent prefix
	got := Run(checks, []string{"cursor"})
	if len(got) != 1 || got[0] != "[cursor] cursor-warn" {
		t.Errorf("expected [[cursor] cursor-warn], got %v", got)
	}
}

func TestRun_perAgentIndependent(t *testing.T) {
	// With two agents selected, a server present for claude only must still
	// emit a warning for cursor.
	checks := []Check{
		{
			Name:   "terraform",
			Agents: nil,
			Probe: func(agent string) bool {
				return agent == "claude" // only claude has it configured
			},
			Warning: "no terraform mcp",
		},
	}
	got := Run(checks, []string{"claude", "cursor"})
	if len(got) != 1 {
		t.Fatalf("expected exactly 1 warning, got %v", got)
	}
	if !strings.HasPrefix(got[0], "[cursor]") {
		t.Errorf("expected warning for cursor, got %q", got[0])
	}
}

func TestRun_multiAgentBothMissing(t *testing.T) {
	checks := []Check{
		{Name: "X", Agents: nil, Probe: func(string) bool { return false }, Warning: "missing"},
	}
	got := Run(checks, []string{"claude", "cursor"})
	if len(got) != 2 {
		t.Fatalf("expected 2 warnings (one per agent), got %v", got)
	}
}

func TestServerKeysFromFile_mcpServers(t *testing.T) {
	dir := t.TempDir()
	writeMCPConfig(t, dir, "settings.json", "mcpServers", []string{"terraform-mcp", "btp-admin"})
	keys := serverKeysFromFile(filepath.Join(dir, "settings.json"))
	want := map[string]bool{"terraform-mcp": true, "btp-admin": true}
	for _, k := range keys {
		delete(want, k)
	}
	if len(want) != 0 {
		t.Errorf("missing keys: %v", want)
	}
}

func TestServerKeysFromFile_servers(t *testing.T) {
	dir := t.TempDir()
	writeMCPConfig(t, dir, "mcp.json", "servers", []string{"opentofu-server"})
	keys := serverKeysFromFile(filepath.Join(dir, "mcp.json"))
	if len(keys) != 1 || keys[0] != "opentofu-server" {
		t.Errorf("unexpected keys: %v", keys)
	}
}

func TestServerKeysFromFile_missing(t *testing.T) {
	if keys := serverKeysFromFile("/nonexistent/path/settings.json"); keys != nil {
		t.Errorf("expected nil for missing file, got %v", keys)
	}
}

func TestTerraformCheck_claudeUserConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	writeMCPConfig(t, home, ".claude.json", "mcpServers", []string{"terraform-mcp"})

	if !terraformCheck.Probe("claude") {
		t.Error("expected terraform check to pass with terraform-mcp in ~/.claude.json")
	}
}

func TestTerraformCheck_claudeProjectConfig(t *testing.T) {
	orig, _ := os.Getwd()
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	home := t.TempDir()
	t.Setenv("HOME", home)

	writeMCPConfig(t, dir, ".mcp.json", "mcpServers", []string{"terraform-mcp"})

	if !terraformCheck.Probe("claude") {
		t.Error("expected terraform check to pass with terraform-mcp in .mcp.json")
	}
}

func TestTerraformCheck_cursorConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	writeMCPConfig(t, filepath.Join(home, ".cursor"), "mcp.json", "mcpServers", []string{"opentofu-server"})

	if !terraformCheck.Probe("cursor") {
		t.Error("expected terraform check to pass with opentofu-server in cursor config")
	}
}

func TestTerraformCheck_cursorProjectConfig(t *testing.T) {
	orig, _ := os.Getwd()
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	home := t.TempDir()
	t.Setenv("HOME", home)

	writeMCPConfig(t, filepath.Join(dir, ".cursor"), "mcp.json", "mcpServers", []string{"opentofu-server"})

	if !terraformCheck.Probe("cursor") {
		t.Error("expected terraform check to pass with opentofu-server in .cursor/mcp.json")
	}
}

func TestTerraformCheck_codexUserConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	writeCodexMCPConfig(t, filepath.Join(home, ".codex", "config.toml"), []string{"terraform-mcp"})

	if !terraformCheck.Probe("codex") {
		t.Error("expected terraform check to pass with terraform-mcp in ~/.codex/config.toml")
	}
}

func TestTerraformCheck_codexProjectConfig(t *testing.T) {
	orig, _ := os.Getwd()
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	home := t.TempDir()
	t.Setenv("HOME", home)

	writeCodexMCPConfig(t, filepath.Join(dir, ".codex", "config.toml"), []string{"opentofu-server"})

	if !terraformCheck.Probe("codex") {
		t.Error("expected terraform check to pass with opentofu-server in .codex/config.toml")
	}
}

func TestTerraformCheck_copilotVSCode(t *testing.T) {
	orig, _ := os.Getwd()
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	writeMCPConfig(t, filepath.Join(dir, ".vscode"), "mcp.json", "servers", []string{"terraform-mcp"})

	if !terraformCheck.Probe("copilot") {
		t.Error("expected terraform check to pass with terraform-mcp in .vscode/mcp.json")
	}
}

func TestTerraformCheck_copilotVSCodeUserLevel(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "xdg"))

	path := vsCodeMCPPath()
	if path == "" {
		t.Skip("vsCodeMCPPath() returned empty (unsupported platform)")
	}
	writeMCPConfig(t, filepath.Dir(path), filepath.Base(path), "servers", []string{"opentofu-server"})

	if !terraformCheck.Probe("copilot") {
		t.Error("expected terraform check to pass with opentofu-server in VS Code user mcp.json")
	}
}

func TestTerraformCheck_mcpDockerNoProfile(t *testing.T) {
	// Gateway configured without --profile: active servers come from registry.yaml.
	home := t.TempDir()
	t.Setenv("HOME", home)

	writeClaudeDockerGateway(t, home, nil) // no --profile arg
	writeRegistryYAML(t, home, []string{"terraform-mcp", "fetch"})

	if !terraformCheck.Probe("claude") {
		t.Error("expected terraform check to pass via registry.yaml when no --profile is set")
	}
}

func TestTerraformCheck_mcpDockerWithProfile(t *testing.T) {
	// Gateway configured with --profile myprofile: active servers come from working_set.
	home := t.TempDir()
	t.Setenv("HOME", home)

	writeClaudeDockerGateway(t, home, []string{"--profile", "myprofile"})
	createDockerMCPDB(t, home, "myprofile", []string{"terraform-mcp", "fetch"})

	if !terraformCheck.Probe("claude") {
		t.Error("expected terraform check to pass via working_set for active profile")
	}
}

func TestTerraformCheck_mcpDockerWithEqualsProfile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	writeClaudeDockerGateway(t, home, []string{"--profile=prod"})
	createDockerMCPDB(t, home, "prod", []string{"terraform-mcp"})

	if !terraformCheck.Probe("claude") {
		t.Error("expected terraform check to pass via --profile=prod")
	}
}

func TestTerraformCheck_mcpDockerInactiveProfileNotSufficient(t *testing.T) {
	// Server in an inactive profile must NOT suppress the warning for the active profile.
	home := t.TempDir()
	t.Setenv("HOME", home)

	writeClaudeDockerGateway(t, home, []string{"--profile", "active"})
	// terraform-mcp only exists in "inactive", not in "active".
	createDockerMCPDB(t, home, "active", []string{"fetch"})
	createDockerMCPDB(t, home, "inactive", []string{"terraform-mcp"})

	if terraformCheck.Probe("claude") {
		t.Error("server in inactive profile must not suppress warning for active profile")
	}
}

func TestTerraformCheck_mcpDockerRegistryWithoutGatewayEntry(t *testing.T) {
	// registry.yaml present but no gateway entry in the client config:
	// Docker is not wired to this client, so it must not count.
	home := t.TempDir()
	t.Setenv("HOME", home)

	// No gateway entry written to ~/.claude.json.
	writeRegistryYAML(t, home, []string{"terraform"})

	if terraformCheck.Probe("claude") {
		t.Error("registry.yaml without a gateway client entry must not suppress the warning")
	}
}

func TestBTPAdminCheck_missing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	orig, _ := os.Getwd()
	dir := t.TempDir()
	_ = os.Chdir(dir)
	t.Cleanup(func() { _ = os.Chdir(orig) })

	for _, agent := range []string{"claude", "codex", "cursor", "copilot"} {
		if btpAdminCheck.Probe(agent) {
			t.Errorf("expected btp check to fail for agent %q when nothing is configured", agent)
		}
	}
}

func TestBTPAdminCheck_present(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	writeMCPConfig(t, home, ".claude.json", "mcpServers", []string{"btp-administration"})

	if !btpAdminCheck.Probe("claude") {
		t.Error("expected btp check to pass with btp-administration in claude settings")
	}
}

func TestBTPAdminCheck_claudeProjectScoped(t *testing.T) {
	orig, _ := os.Getwd()
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	resolvedDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		resolvedDir = dir
	}

	home := t.TempDir()
	t.Setenv("HOME", home)

	claudeJSON := map[string]any{
		"projects": map[string]any{
			resolvedDir: map[string]any{
				"mcpServers": map[string]any{
					"BTP-Administration": map[string]string{"type": "http", "url": "https://example.com"},
				},
			},
		},
	}
	data, err := json.Marshal(claudeJSON)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".claude.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	if !btpAdminCheck.Probe("claude") {
		t.Error("expected btp check to pass with BTP-Administration in ~/.claude.json projects scope")
	}
}

func TestBTPAdminCheck_claudeOnlyDoesNotSatisfyCursor(t *testing.T) {
	// BTP configured for claude must not suppress the warning for cursor.
	home := t.TempDir()
	t.Setenv("HOME", home)

	orig, _ := os.Getwd()
	dir := t.TempDir()
	_ = os.Chdir(dir)
	t.Cleanup(func() { _ = os.Chdir(orig) })

	writeMCPConfig(t, home, ".claude.json", "mcpServers", []string{"btp-administration"})

	got := Run([]Check{btpAdminCheck}, []string{"claude", "cursor"})
	if len(got) != 1 {
		t.Fatalf("expected exactly 1 warning (for cursor), got %v", got)
	}
	if !strings.HasPrefix(got[0], "[cursor]") {
		t.Errorf("expected warning for cursor, got %q", got[0])
	}
}

func TestBTPAdminCheck_codexPresent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	writeCodexMCPConfig(t, filepath.Join(home, ".codex", "config.toml"), []string{"btp-administration"})

	if !btpAdminCheck.Probe("codex") {
		t.Error("expected btp check to pass with btp-administration in ~/.codex/config.toml")
	}
}

func TestCodexServerNamesFromTOMLIgnoresNestedTables(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	data := strings.Join([]string{
		`[mcp_servers.terraform-mcp]`,
		`command = "terraform-mcp"`,
		``,
		`[mcp_servers.terraform-mcp.env]`,
		`TOKEN = "x"`,
		``,
		`[mcp_servers."btp-administration"]`,
		`command = "btp-admin"`,
	}, "\n")
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	got := codexServerNamesFromTOML(path)
	if len(got) != 2 {
		t.Fatalf("expected 2 codex MCP servers, got %v", got)
	}
	if !hasMatch(got, "terraform") || !hasMatch(got, "btp") {
		t.Fatalf("expected terraform and btp servers, got %v", got)
	}
}

// writeClaudeDockerGateway writes a Docker MCP gateway entry into ~/.claude.json.
// extraArgs are appended after ["mcp", "gateway", "run"] — pass nil for no --profile.
func writeClaudeDockerGateway(t *testing.T, home string, extraArgs []string) {
	t.Helper()
	args := append([]string{"mcp", "gateway", "run"}, extraArgs...)
	cfg := map[string]any{
		"mcpServers": map[string]any{
			"MCP_DOCKER": map[string]any{"command": "docker", "args": args},
		},
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".claude.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeCodexMCPConfig(t *testing.T, path string, servers []string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for i, server := range servers {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("[mcp_servers.")
		b.WriteString(server)
		b.WriteString("]\ncommand = \"")
		b.WriteString(server)
		b.WriteString("\"\n")
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeRegistryYAML writes a registry.yaml with the given server names.
func writeRegistryYAML(t *testing.T, home string, servers []string) {
	t.Helper()
	dir := filepath.Join(home, ".docker", "mcp")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	b.WriteString("registry:\n")
	for _, s := range servers {
		b.WriteString("  ")
		b.WriteString(s)
		b.WriteString(":\n    ref: \"\"\n")
	}
	if err := os.WriteFile(filepath.Join(dir, "registry.yaml"), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

// createDockerMCPDB inserts a profile row into mcp-toolkit.db using the
// pure-Go SQLite driver (same driver as production code).
func createDockerMCPDB(t *testing.T, home, profileID string, serverNames []string) {
	t.Helper()
	dbDir := filepath.Join(home, ".docker", "mcp")
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(dbDir, "mcp-toolkit.db")

	entries := make([]map[string]any, len(serverNames))
	for i, name := range serverNames {
		entries[i] = map[string]any{"snapshot": map[string]any{"server": map[string]any{"name": name}}}
	}
	serversJSON, err := json.Marshal(entries)
	if err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlite", "file:"+dbPath+"?mode=rwc")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS working_set (id text primary key, name text not null, servers text not null)`); err != nil {
		t.Fatal(err)
	}
	// INSERT OR REPLACE so multiple profiles can coexist in one DB.
	if _, err := db.Exec(`INSERT OR REPLACE INTO working_set VALUES (?,?,?)`, profileID, profileID+" Profile", string(serversJSON)); err != nil {
		t.Fatal(err)
	}
}
