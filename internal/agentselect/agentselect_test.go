package agentselect

import "testing"

func TestFromFlagSupportsCodex(t *testing.T) {
	agents, err := fromFlag("claude,codex")
	if err != nil {
		t.Fatalf("fromFlag: %v", err)
	}
	if len(agents) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(agents))
	}
	if agents[0].ID != "claude" || agents[1].ID != "codex" {
		t.Fatalf("unexpected agents: %#v", agents)
	}
}

func TestFromFlagUnknownAgentListsCodex(t *testing.T) {
	_, err := fromFlag("unknown")
	if err == nil {
		t.Fatal("expected error for unknown agent")
	}
	if got := err.Error(); got != `unknown agent "unknown" — supported: claude, codex, cursor, copilot` {
		t.Fatalf("unexpected error: %q", got)
	}
}
