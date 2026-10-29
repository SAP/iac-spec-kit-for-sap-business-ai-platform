package agentselect

import (
	"slices"
	"strings"
	"testing"
)

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
	if got := err.Error(); got != `unknown agent "unknown" — supported: claude, codex, cursor, copilot, gemini` {
		t.Fatalf("unexpected error: %q", got)
	}
}

func TestGeminiSelection(t *testing.T) {
	for _, flag := range []string{"gemini", "claude, gemini", "claude,codex,cursor,copilot,gemini"} {
		agents, err := fromFlag(flag)
		if err != nil {
			t.Fatal(err)
		}
		if len(agents) != len(strings.Split(flag, ",")) || agents[len(agents)-1].ID != "gemini" {
			t.Fatalf("unexpected selection for %q: %v", flag, agents)
		}
	}
	if !slices.Contains(agentOrder, "gemini") {
		t.Fatal("Gemini missing from interactive options")
	}
	_, err := fromFlag(" ")
	if err == nil || !strings.Contains(err.Error(), "gemini") {
		t.Fatalf("empty selection error must list Gemini: %v", err)
	}
}
