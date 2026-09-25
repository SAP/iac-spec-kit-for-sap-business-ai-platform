package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/SAP/btp-iac-spec-kit/internal/eval"
)

func main() {
	provider := flag.String("provider", "all", "codex, claude, or all")
	specPath := flag.String("eval", "skills/sap-iac-govern/evals/evals.json", "eval specification")
	skillPath := flag.String("skill", "skills/sap-iac-govern/SKILL.md", "skill instructions")
	judgeName := flag.String("judge", "codex", "codex or claude")
	keep := flag.Bool("keep", false, "retain fixture directories")
	live := flag.Bool("live", false, "allow authenticated live agent calls")
	timeout := flag.Duration("timeout", 30*time.Minute, "maximum duration for each provider run")
	caseTimeout := flag.Duration("case-timeout", 5*time.Minute, "maximum duration for one eval case, including judging")
	flag.Parse()
	if !*live {
		fmt.Fprintln(os.Stderr, "refusing live model calls: pass -live explicitly")
		os.Exit(2)
	}
	spec, err := eval.Load(*specPath)
	if err != nil {
		fail(err)
	}
	skill, err := os.ReadFile(*skillPath)
	if err != nil {
		fail(err)
	}
	judge, err := agent(*judgeName)
	if err != nil {
		fail(err)
	}
	providers := []string{*provider}
	if *provider == "all" {
		providers = []string{"codex", "claude"}
	}
	failed := false
	for _, name := range providers {
		a, err := agent(name)
		if err != nil {
			fail(err)
		}
		if name == judge.Name() {
			fmt.Fprintf(os.Stderr, "warning: provider and judge are both %s; use a different judge to reduce self-judging bias\n", name)
		}
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		report, err := eval.RunWithOptions(ctx, spec, string(skill), a, judge, *keep, eval.Options{CaseTimeout: *caseTimeout, Retries: 1})
		cancel()
		if err != nil {
			fail(err)
		}
		report.Judge = judge.Name()
		report.GeneratedAt = time.Now().UTC()
		report.GitSHA = commandOutput("git", "rev-parse", "HEAD")
		report.SpecPath = *specPath
		report.SkillPath = *skillPath
		report.AgentVersion = commandOutput(name, "--version")
		report.JudgeVersion = commandOutput(*judgeName, "--version")
		b, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			fail(err)
		}
		path := filepath.Join("eval-artifacts", name+"-report.json")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			fail(err)
		}
		if err := os.WriteFile(path, b, 0o644); err != nil {
			fail(err)
		}
		fmt.Println(path)
		if report.Failed() {
			fmt.Fprintf(os.Stderr, "%s: one or more eval cases failed; see %s\n", name, path)
			failed = true
		}
	}
	if failed {
		os.Exit(1)
	}
}
func agent(name string) (eval.Agent, error) {
	switch name {
	case "codex":
		return eval.Codex{}, nil
	case "claude":
		return eval.Claude{}, nil
	default:
		return nil, fmt.Errorf("unknown provider %q", name)
	}
}
func fail(err error) { fmt.Fprintln(os.Stderr, "sap-iac-eval:", err); os.Exit(1) }

func commandOutput(name string, args ...string) string {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
