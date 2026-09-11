// Package eval runs skill eval specifications against coding-agent CLIs.
package eval

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type File struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}
type Turn struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type Environment struct {
	WorkingDirectory          string `json:"working_directory"`
	ForbidAncestorProjectRoot bool   `json:"forbid_ancestor_project_root"`
}
type Assertion struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// Case deliberately does not decode the optional JSON trap field: it is eval-author
// documentation and must never be supplied to the agent or judge.
type Case struct {
	ID          int         `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Prompt      string      `json:"prompt"`
	Files       []File      `json:"files"`
	Turns       []Turn      `json:"turns"`
	Environment Environment `json:"environment"`
	Assertions  []Assertion `json:"assertions"`
}
type Spec struct {
	SkillName string `json:"skill_name"`
	Evals     []Case `json:"evals"`
}

func Load(path string) (Spec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Spec{}, err
	}
	var s Spec
	if err := json.Unmarshal(b, &s); err != nil {
		return Spec{}, fmt.Errorf("parse eval spec: %w", err)
	}
	if s.SkillName == "" || len(s.Evals) == 0 {
		return Spec{}, fmt.Errorf("eval spec must contain skill_name and evals")
	}
	caseIDs := map[int]bool{}
	for _, c := range s.Evals {
		if c.ID == 0 || c.Name == "" || len(c.Assertions) == 0 {
			return Spec{}, fmt.Errorf("eval case must contain id, name, and assertions")
		}
		if caseIDs[c.ID] {
			return Spec{}, fmt.Errorf("duplicate eval case id %d", c.ID)
		}
		caseIDs[c.ID] = true
		assertionIDs := map[string]bool{}
		for _, a := range c.Assertions {
			if strings.TrimSpace(a.ID) == "" || strings.TrimSpace(a.Text) == "" {
				return Spec{}, fmt.Errorf("case %d has assertion with empty id or text", c.ID)
			}
			if assertionIDs[a.ID] {
				return Spec{}, fmt.Errorf("case %d has duplicate assertion id %q", c.ID, a.ID)
			}
			assertionIDs[a.ID] = true
		}
		if len(c.Turns) == 0 && strings.TrimSpace(c.Prompt) == "" {
			return Spec{}, fmt.Errorf("case %d must contain prompt or turns", c.ID)
		}
		if len(c.Turns) > 0 && c.Prompt != "" && c.Prompt != c.Turns[0].Content {
			return Spec{}, fmt.Errorf("case %d prompt must match first turn when both are supplied", c.ID)
		}
	}
	return s, nil
}

// Agent is intentionally CLI-oriented: authenticated vendor CLIs remain the credential boundary.
type Agent interface {
	Name() string
	SupportsResume() bool
	Turn(context.Context, string, string, string) (response, session string, err error)
}

type judgeAgent interface {
	JudgeTurn(context.Context, string, string, string) (response, session string, err error)
}

type Report struct {
	Provider     string       `json:"provider"`
	Judge        string       `json:"judge,omitempty"`
	GeneratedAt  time.Time    `json:"generated_at,omitempty"`
	GitSHA       string       `json:"git_sha,omitempty"`
	SpecPath     string       `json:"spec_path,omitempty"`
	SkillPath    string       `json:"skill_path,omitempty"`
	AgentVersion string       `json:"agent_version,omitempty"`
	JudgeVersion string       `json:"judge_version,omitempty"`
	Cases        []CaseResult `json:"cases"`
}
type CaseResult struct {
	ID               int                 `json:"id"`
	Name             string              `json:"name"`
	Directory        string              `json:"directory"`
	Transcript       []Message           `json:"transcript"`
	InitialArtifacts map[string]string   `json:"initial_artifacts"`
	TurnArtifacts    []map[string]string `json:"turn_artifacts"`
	Artifacts        map[string]string   `json:"artifacts"`
	Assertions       []Assertion         `json:"assertions"`
	Judge            []Verdict           `json:"judge"`
	DurationMS       int64               `json:"duration_ms"`
	Error            string              `json:"error,omitempty"`
}
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type Verdict struct {
	ID     string `json:"id"`
	Passed bool   `json:"passed"`
	Reason string `json:"reason"`
}

// Failed reports whether an eval could not run or an assertion did not pass.
func (r Report) Failed() bool {
	for _, c := range r.Cases {
		if c.Error != "" {
			return true
		}
		expected := map[string]bool{}
		for _, assertion := range c.Assertions {
			expected[assertion.ID] = true
		}
		seen := map[string]bool{}
		for _, verdict := range c.Judge {
			if !expected[verdict.ID] || seen[verdict.ID] {
				return true
			}
			seen[verdict.ID] = true
			if !verdict.Passed {
				return true
			}
		}
		for _, assertion := range c.Assertions {
			if !seen[assertion.ID] {
				return true
			}
		}
	}
	return false
}

// Run materializes each case, performs all turns, then asks judge to score its assertions.
// Directories are retained when keep is true, which makes failures inspectable.
type Options struct {
	CaseTimeout time.Duration
	Retries     int
}

func Run(ctx context.Context, spec Spec, skill string, agent, judge Agent, keep bool) (Report, error) {
	return RunWithOptions(ctx, spec, skill, agent, judge, keep, Options{Retries: 1})
}

func RunWithOptions(ctx context.Context, spec Spec, skill string, agent, judge Agent, keep bool, options Options) (Report, error) {
	r := Report{Provider: agent.Name()}
	for _, c := range spec.Evals {
		caseStart := time.Now()
		caseCtx := ctx
		cancel := func() {}
		if options.CaseTimeout > 0 {
			caseCtx, cancel = context.WithTimeout(ctx, options.CaseTimeout)
		}
		result := CaseResult{ID: c.ID, Name: c.Name, Assertions: c.Assertions}
		dir, fixtureRoot, err := fixture(c)
		if err != nil {
			cancel()
			return r, err
		}
		isolate := c.Environment.ForbidAncestorProjectRoot || c.Environment.WorkingDirectory == "isolated-no-project-root"
		if c.Environment.WorkingDirectory != "" && c.Environment.WorkingDirectory != "isolated-no-project-root" {
			cancel()
			_ = os.RemoveAll(fixtureRoot)
			return r, fmt.Errorf("unsupported eval working_directory %q", c.Environment.WorkingDirectory)
		}
		if isolate {
			if err := requireNoProjectRootAncestor(dir); err != nil {
				cancel()
				_ = os.RemoveAll(fixtureRoot)
				return r, err
			}
		}
		result.Directory = dir
		result.InitialArtifacts, err = snapshot(dir)
		if err != nil {
			cancel()
			_ = os.RemoveAll(fixtureRoot)
			return r, err
		}
		turns := c.Turns
		if len(turns) == 0 {
			turns = []Turn{{Role: "user", Content: c.Prompt}}
		}
		var session string
		for i, turn := range turns {
			prompt := turn.Content
			if i == 0 || !agent.SupportsResume() {
				// Resumable providers retain this first-turn context; stateless providers receive it on every turn.
				if i > 0 {
					prompt = "Conversation so far:\n" + transcriptText(result.Transcript) + "\n\nLatest user request:\n" + prompt
				}
				prompt = "Follow this skill exactly and work only in the current project directory.\n\n" + skill + "\n\n" + prompt
			}
			result.Transcript = append(result.Transcript, Message(turn))
			response, next, runErr := turnWithRetry(caseCtx, agent, dir, session, prompt, options.Retries)
			if runErr != nil && session != "" && agent.SupportsResume() {
				// A lost provider session should not discard an otherwise valid multi-turn eval.
				fallback := "Follow this skill exactly and work only in the current project directory.\n\n" + skill + "\n\nConversation so far:\n" + transcriptText(result.Transcript)
				response, next, runErr = turnWithRetry(caseCtx, agent, dir, "", fallback, options.Retries)
			}
			if runErr != nil {
				result.Error = runErr.Error()
				break
			}
			session = next
			if !agent.SupportsResume() {
				session = ""
			}
			result.Transcript = append(result.Transcript, Message{Role: "assistant", Content: response})
			if turnArtifacts, snapshotErr := snapshot(dir); snapshotErr != nil {
				result.Error = snapshotErr.Error()
				break
			} else {
				result.TurnArtifacts = append(result.TurnArtifacts, turnArtifacts)
			}
		}
		result.Artifacts, err = snapshot(dir)
		if err != nil {
			result.Error = err.Error()
		}
		if result.Error == "" {
			result.Judge, err = scoreWithRetry(caseCtx, judge, c.Assertions, result, options.Retries)
			if err != nil {
				result.Error = err.Error()
			}
		}
		r.Cases = append(r.Cases, result)
		cancel()
		r.Cases[len(r.Cases)-1].DurationMS = time.Since(caseStart).Milliseconds()
		if !keep {
			_ = os.RemoveAll(fixtureRoot)
		}
	}
	return r, nil
}

func turnWithRetry(ctx context.Context, agent Agent, dir, session, prompt string, retries int) (string, string, error) {
	response, next, err := agent.Turn(ctx, dir, session, prompt)
	if err != nil && retries > 0 && ctx.Err() == nil {
		return agent.Turn(ctx, dir, session, prompt)
	}
	return response, next, err
}

func transcriptText(messages []Message) string {
	var b strings.Builder
	for _, message := range messages {
		fmt.Fprintf(&b, "%s: %s\n", message.Role, message.Content)
	}
	return b.String()
}

func fixture(c Case) (dir, root string, err error) {
	root, err = os.MkdirTemp("", "btp-iac-eval-")
	if err != nil {
		return "", "", err
	}
	dir = root
	if c.Environment.WorkingDirectory != "" {
		if c.Environment.WorkingDirectory != "isolated-no-project-root" {
			_ = os.RemoveAll(root)
			return "", "", fmt.Errorf("unsupported eval working_directory %q", c.Environment.WorkingDirectory)
		}
		dir = filepath.Join(root, c.Environment.WorkingDirectory)
		if err := os.Mkdir(dir, 0o755); err != nil {
			_ = os.RemoveAll(root)
			return "", "", err
		}
	}
	for _, f := range c.Files {
		p := filepath.Join(dir, filepath.Clean(f.Path))
		if !strings.HasPrefix(p, dir+string(os.PathSeparator)) {
			_ = os.RemoveAll(root)
			return "", "", fmt.Errorf("unsafe fixture path %q", f.Path)
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			_ = os.RemoveAll(root)
			return "", "", err
		}
		if err := os.WriteFile(p, []byte(f.Content), 0o644); err != nil {
			_ = os.RemoveAll(root)
			return "", "", err
		}
	}
	return dir, root, nil
}

func snapshot(root string) (map[string]string, error) {
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		out[filepath.ToSlash(rel)] = string(b)
		return nil
	})
	return out, err
}

func requireNoProjectRootAncestor(dir string) error {
	for current := dir; ; current = filepath.Dir(current) {
		if hasProjectMarker(current) {
			return fmt.Errorf("isolated eval directory %q has project-root ancestor %q", dir, current)
		}
		parent := filepath.Dir(current)
		if parent == current {
			return nil
		}
	}
}

// Keep this predicate aligned with scaffold.IsProject without making the evaluator
// depend on scaffold's embedded command assets.
func hasProjectMarker(dir string) bool {
	for _, name := range []string{"specs", "memory"} {
		if info, err := os.Stat(filepath.Join(dir, name)); err == nil && info.IsDir() {
			return true
		}
	}
	return false
}

func score(ctx context.Context, judge Agent, assertions []Assertion, result CaseResult) ([]Verdict, error) {
	artifacts := judgeArtifacts(result.Artifacts)
	b, _ := json.Marshal(struct {
		Assertions       []Assertion         `json:"assertions"`
		Transcript       []Message           `json:"transcript"`
		InitialArtifacts map[string]string   `json:"initial_artifacts"`
		TurnArtifacts    []map[string]string `json:"turn_artifacts"`
		Artifacts        map[string]string   `json:"artifacts"`
	}{assertions, result.Transcript, judgeArtifacts(result.InitialArtifacts), judgeTurnArtifacts(result.TurnArtifacts), artifacts})
	prompt := "You are an impartial software-eval judge. Judge each assertion only from its literal current text and the evidence below; never substitute an earlier or stricter version. If an assertion explicitly names an artifact, decide it from that artifact only, not the assistant's prose. For temporal assertions, use turn_artifacts, where element N is the filesystem after assistant turn N. For an assertion naming initial_artifacts and artifacts, compare those two supplied values directly. Do not impose exact wording, a single question mark, or a single sentence unless the assertion explicitly requires it. Return JSON only: {\"verdicts\":[{\"id\":string,\"passed\":boolean,\"reason\":string}]}.\nEvidence:\n" + string(b)
	judgeDir, err := os.MkdirTemp("", "btp-iac-eval-judge-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(judgeDir) }()
	var text string
	if isolatedJudge, ok := judge.(judgeAgent); ok {
		text, _, err = isolatedJudge.JudgeTurn(ctx, judgeDir, "", prompt)
	} else {
		text, _, err = judge.Turn(ctx, judgeDir, "", prompt)
	}
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Verdicts []Verdict `json:"verdicts"`
	}
	if err := json.Unmarshal([]byte(extractJSON(text)), &parsed); err != nil {
		return nil, fmt.Errorf("parse judge response: %w", err)
	}
	return validateVerdicts(assertions, parsed.Verdicts)
}

func scoreWithRetry(ctx context.Context, judge Agent, assertions []Assertion, result CaseResult, retries int) ([]Verdict, error) {
	verdicts, err := score(ctx, judge, assertions, result)
	if err != nil && retries > 0 && ctx.Err() == nil {
		return score(ctx, judge, assertions, result)
	}
	return verdicts, err
}

func validateVerdicts(assertions []Assertion, verdicts []Verdict) ([]Verdict, error) {
	expected := make(map[string]bool, len(assertions))
	for _, assertion := range assertions {
		expected[assertion.ID] = true
	}
	seen := make(map[string]bool, len(verdicts))
	for _, verdict := range verdicts {
		if !expected[verdict.ID] {
			return nil, fmt.Errorf("judge returned unknown assertion id %q", verdict.ID)
		}
		if seen[verdict.ID] {
			return nil, fmt.Errorf("judge returned duplicate verdict for assertion %q", verdict.ID)
		}
		seen[verdict.ID] = true
	}
	for _, assertion := range assertions {
		if !seen[assertion.ID] {
			verdicts = append(verdicts, Verdict{ID: assertion.ID, Passed: false, Reason: "judge omitted verdict"})
		}
	}
	return verdicts, nil
}
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if json.Valid([]byte(s)) {
		return s
	}
	i := strings.Index(s, "{")
	if i >= 0 {
		var raw json.RawMessage
		if err := json.NewDecoder(strings.NewReader(s[i:])).Decode(&raw); err == nil {
			return string(raw)
		}
	}
	return s
}

const maxJudgeArtifactBytesPerArtifact = 12_000

func judgeArtifacts(artifacts map[string]string) map[string]string {
	bounded := make(map[string]string, len(artifacts))
	for _, path := range SortedArtifacts(artifacts) {
		value := artifacts[path]
		if len(value) > maxJudgeArtifactBytesPerArtifact {
			bounded[path] = value[:maxJudgeArtifactBytesPerArtifact] + "\n<truncated for judge>"
			continue
		}
		bounded[path] = value
	}
	return bounded
}

func judgeTurnArtifacts(snapshots []map[string]string) []map[string]string {
	bounded := make([]map[string]string, len(snapshots))
	for i, snapshot := range snapshots {
		bounded[i] = judgeArtifacts(snapshot)
	}
	return bounded
}

type Codex struct{}

func (Codex) Name() string         { return "codex" }
func (Codex) SupportsResume() bool { return true }
func (Codex) Turn(ctx context.Context, dir, session, prompt string) (string, string, error) {
	if _, err := exec.LookPath("codex"); err != nil {
		return "", "", err
	}
	outFile, err := os.CreateTemp("", "btp-iac-eval-codex-message-")
	if err != nil {
		return "", session, err
	}
	out := outFile.Name()
	if err := outFile.Close(); err != nil {
		_ = os.Remove(out)
		return "", session, err
	}
	defer func() { _ = os.Remove(out) }()
	var args []string
	if session != "" {
		args = []string{"exec", "resume", "--skip-git-repo-check", "--json", "-o", out, session, prompt}
	} else {
		args = []string{"exec", "--skip-git-repo-check", "--sandbox", "workspace-write", "-C", dir, "--json", "-o", out, prompt}
	}
	cmd := exec.CommandContext(ctx, "codex", args...)
	cmd.Dir = dir
	raw, err := cmd.CombinedOutput()
	if err != nil {
		return "", session, fmt.Errorf("codex: %w: %s", err, strings.TrimSpace(string(raw)))
	}
	b, readErr := os.ReadFile(out)
	if readErr != nil {
		return "", session, readErr
	}
	return string(b), threadID(raw, session), nil
}

type Claude struct{}

func (Claude) Name() string         { return "claude" }
func (Claude) SupportsResume() bool { return true }
func (Claude) Turn(ctx context.Context, dir, session, prompt string) (string, string, error) {
	return Claude{}.turn(ctx, dir, session, prompt, false)
}

// JudgeTurn disables user customizations and tools so judging cannot be affected by local hooks.
func (Claude) JudgeTurn(ctx context.Context, dir, session, prompt string) (string, string, error) {
	return Claude{}.turn(ctx, dir, session, prompt, true)
}

func (Claude) turn(ctx context.Context, dir, session, prompt string, judge bool) (string, string, error) {
	if _, err := exec.LookPath("claude"); err != nil {
		return "", "", err
	}
	args := []string{"-p", "--output-format", "json", "--permission-mode", "acceptEdits"}
	if judge {
		args = append(args, "--safe-mode", "--tools", "")
	}
	if session != "" {
		args = append(args, "--resume", session)
	}
	args = append(args, prompt)
	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	raw, err := cmd.Output()
	if err != nil {
		return "", session, fmt.Errorf("claude: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	var v struct {
		SessionID string `json:"session_id"`
		Result    string `json:"result"`
	}
	if err := json.Unmarshal([]byte(extractJSON(string(raw))), &v); err != nil {
		return "", session, err
	}
	if v.Result == "" {
		return "", session, fmt.Errorf("claude returned no result")
	}
	if session == "" && !judge && v.SessionID == "" {
		return "", session, fmt.Errorf("claude returned no session_id for a resumable run")
	}
	return v.Result, v.SessionID, nil
}

func threadID(raw []byte, fallback string) string {
	for _, line := range strings.Split(string(raw), "\n") {
		var v map[string]any
		if json.Unmarshal([]byte(line), &v) == nil {
			if id := findID(v); id != "" {
				return id
			}
		}
	}
	return fallback
}
func findID(v any) string {
	switch x := v.(type) {
	case map[string]any:
		for _, k := range []string{"thread_id", "session_id", "id"} {
			if s, ok := x[k].(string); ok && strings.Contains(s, "-") {
				return s
			}
		}
		for _, child := range x {
			if id := findID(child); id != "" {
				return id
			}
		}
	case []any:
		for _, child := range x {
			if id := findID(child); id != "" {
				return id
			}
		}
	}
	return ""
}

func SortedArtifacts(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
