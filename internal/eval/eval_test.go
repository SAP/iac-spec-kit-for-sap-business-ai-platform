package eval

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndFixture(t *testing.T) {
	cases := []struct {
		skill string
		paths []string
	}{
		{"sap-iac-govern", []string{"specs/.gitkeep", "memory/.gitkeep", "terraform/.gitkeep"}},
		{"sap-iac-accounts", []string{"specs/scenario.md", "memory/governance.md", "terraform/.gitkeep"}},
		{"sap-iac-security", []string{"specs/scenario.md", "specs/landscape.md", "memory/.gitkeep", "terraform/.gitkeep"}},
		{"sap-iac-tasks", []string{"specs/landscape.md", "specs/services.md", "specs/trust.md"}},
		{"sap-iac-generate", []string{"specs/tasks.md", "terraform/.gitkeep"}},
	}
	for _, tc := range cases {
		t.Run(tc.skill, func(t *testing.T) {
			s, err := Load(filepath.Join("..", "..", "skills", tc.skill, "evals", "evals.json"))
			if err != nil {
				t.Fatal(err)
			}
			if len(s.Evals) == 0 {
				t.Fatal("expected at least one eval")
			}
			d, root, err := fixture(s.Evals[0])
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := os.RemoveAll(root); err != nil {
					t.Errorf("remove fixture root: %v", err)
				}
			})
			for _, p := range tc.paths {
				if _, err := os.Stat(filepath.Join(d, p)); err != nil {
					t.Errorf("missing fixture %s: %v", p, err)
				}
			}
		})
	}
}
func TestExtractJSON(t *testing.T) {
	got := extractJSON("note {\"verdicts\":[]} trailing")
	if got != "{\"verdicts\":[]}" {
		t.Fatalf("got %q", got)
	}
}

func TestRequireNoProjectRootAncestor(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"specs", "memory", "terraform"} {
		if err := os.Mkdir(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	nested := filepath.Join(root, "nested")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := requireNoProjectRootAncestor(nested); err == nil {
		t.Fatal("expected project-root ancestor error")
	}
}

func TestReportFailed(t *testing.T) {
	if (Report{}).Failed() {
		t.Fatal("empty report should pass")
	}
	if !(Report{Cases: []CaseResult{{Error: "provider failed"}}}).Failed() {
		t.Fatal("case error should fail report")
	}
	if !(Report{Cases: []CaseResult{{Judge: []Verdict{{ID: "1", Passed: false}}}}}).Failed() {
		t.Fatal("failed verdict should fail report")
	}
}

func TestLoadRejectsInvalidAssertions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "evals.json")
	data := "{\"skill_name\":\"x\",\"evals\":[{\"id\":1,\"name\":\"one\",\"prompt\":\"go\",\"assertions\":[{\"id\":\"1.1\",\"text\":\"a\"},{\"id\":\"1.1\",\"text\":\"b\"}]}]}"
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected duplicate assertion validation error")
	}
}

func TestValidateVerdicts(t *testing.T) {
	assertions := []Assertion{{ID: "1.1", Text: "one"}, {ID: "1.2", Text: "two"}}
	got, err := validateVerdicts(assertions, []Verdict{{ID: "1.1", Passed: true}})
	if err != nil || len(got) != 2 || got[1].Passed || got[1].Reason != "judge omitted verdict" {
		t.Fatalf("got %#v, %v", got, err)
	}
	if _, err := validateVerdicts(assertions, []Verdict{{ID: "x"}}); err == nil {
		t.Fatal("expected unknown ID error")
	}
	if _, err := validateVerdicts(assertions, []Verdict{{ID: "1.1"}, {ID: "1.1"}}); err == nil {
		t.Fatal("expected duplicate ID error")
	}
}

type fakeAgent struct {
	resume    bool
	calls     int
	failFirst bool
	prompts   []string
}

func (a *fakeAgent) Name() string         { return "fake" }
func (a *fakeAgent) SupportsResume() bool { return a.resume }
func (a *fakeAgent) Turn(_ context.Context, dir, _ string, prompt string) (string, string, error) {
	a.calls++
	a.prompts = append(a.prompts, prompt)
	if a.failFirst && a.calls == 1 {
		return "", "", errors.New("temporary")
	}
	if err := os.WriteFile(filepath.Join(dir, "created.txt"), []byte(string(rune('0'+a.calls))), 0o644); err != nil {
		return "", "", err
	}
	return "done", "session", nil
}

type fakeJudge struct{ fakeAgent }

func (a *fakeJudge) JudgeTurn(_ context.Context, _ string, _ string, _ string) (string, string, error) {
	return "{\"verdicts\":[{\"id\":\"1.1\",\"passed\":true,\"reason\":\"ok\"}]}", "", nil
}

func TestRunSnapshotsTurnsAndRetries(t *testing.T) {
	a := &fakeAgent{failFirst: true}
	j := &fakeJudge{}
	s := Spec{SkillName: "x", Evals: []Case{{ID: 1, Name: "x", Prompt: "first", Turns: []Turn{{Role: "user", Content: "first"}, {Role: "user", Content: "second"}}, Assertions: []Assertion{{ID: "1.1", Text: "ok"}}}}}
	r, err := RunWithOptions(context.Background(), s, "skill", a, j, false, Options{Retries: 1})
	if err != nil {
		t.Fatal(err)
	}
	if r.Failed() || a.calls != 3 || len(r.Cases[0].TurnArtifacts) != 2 {
		t.Fatalf("report=%#v calls=%d", r, a.calls)
	}
	if r.Cases[0].TurnArtifacts[0]["created.txt"] != "2" {
		t.Fatalf("first snapshot %#v", r.Cases[0].TurnArtifacts[0])
	}
}

func TestJudgePayloadUsesSnakeCase(t *testing.T) {
	b, err := json.Marshal(Message{Role: "user", Content: "x"})
	if err != nil || string(b) != "{\"role\":\"user\",\"content\":\"x\"}" {
		t.Fatalf("%s %v", b, err)
	}
}
