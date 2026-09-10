package eval

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndFixture(t *testing.T) {
	s, err := Load(filepath.Join("..", "..", "skills", "btp-iac-govern", "evals", "evals.json"))
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
	defer os.RemoveAll(root)
	for _, p := range []string{"specs/.gitkeep", "memory/.gitkeep", "terraform/.gitkeep"} {
		if _, err := os.Stat(filepath.Join(d, p)); err != nil {
			t.Errorf("missing fixture %s: %v", p, err)
		}
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
