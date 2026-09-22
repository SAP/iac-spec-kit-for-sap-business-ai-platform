package skills

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

const btpReadOnlyBoundary = "When invoking the BTP CLI or any BTP MCP tool, perform **only read or list retrievals**."

// noBTPBoundary lists skills that perform no BTP operations and are exempt from BTP boundary assertions.
var noBTPBoundary = map[string]bool{
	"btp-iac-next/SKILL.md": true,
}

func TestAllSkillsContainPlatformValidationContract(t *testing.T) {
	err := fs.WalkDir(Commands, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, "/SKILL.md") {
			return nil
		}
		if noBTPBoundary[strings.TrimPrefix(path, "./")] {
			return nil
		}
		data, err := Commands.ReadFile(path)
		if err != nil {
			return err
		}
		if !strings.Contains(string(data), ".btp-iac/platform-validation.md") {
			t.Errorf("%s is missing platform validation guidance", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAllSkillsEnforceBTPReadOnlyBoundary(t *testing.T) {
	forbiddenCLICommand := regexp.MustCompile(`(?m)(?:^|[\s` + "`" + `])btp\s+(?:create|update|delete|assign|unassign|enable|disable|set|unset|login|logout|config|profile)\b`)
	err := fs.WalkDir(Commands, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, "/SKILL.md") {
			return nil
		}
		if noBTPBoundary[strings.TrimPrefix(path, "./")] {
			return nil
		}

		data, err := Commands.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(data)
		for _, requirement := range []string{
			btpReadOnlyBoundary,
			"BTP MCP, invoke only a tool explicitly documented as a read/list lookup.",
			"`btp target --global-account <subdomain>` is the sole permitted account-selection prelude",
			"Never invoke, suggest, or approve a BTP operation that creates, updates, deletes",
		} {
			if !strings.Contains(content, requirement) {
				t.Errorf("%s is missing BTP read-only requirement %q", path, requirement)
			}
		}
		if strings.Contains(content, "MCO") {
			t.Errorf("%s refers to MCO; BTP MCP is the supported integration", path)
		}
		if forbiddenCLICommand.MatchString(content) {
			t.Errorf("%s contains a forbidden mutating BTP CLI command", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRequestedSkillsContainSpecificPlatformChecks(t *testing.T) {
	checks := map[string][]string{
		"btp-iac-govern/SKILL.md":   {"btp list accounts/entitlement", "available-region", "NEO", "AWS", "Microsoft Azure", "Google Cloud", "SAP Cloud Infrastructure", "Alibaba Cloud", "Preferred infrastructure provider"},
		"btp-iac-scenario/SKILL.md": {"btp list accounts/entitlement", "available-region", "NEO", "Preferred infrastructure provider"},
		"btp-iac-accounts/SKILL.md": {"btp list accounts/entitlement", "available-region", "NEO", "Preferred infrastructure provider"},
		"btp-iac-generate/SKILL.md": {"available-region", "NEO", "Preferred infrastructure provider"},
		"btp-iac-services/SKILL.md": {"btp list accounts/entitlement"},
	}
	for path, wants := range checks {
		data, err := Commands.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range wants {
			if !strings.Contains(string(data), want) {
				t.Errorf("%s is missing %q", path, want)
			}
		}
	}
}

func TestGreenfieldSkillsHaveSingleNextStepFooter(t *testing.T) {
	linearFlowSkills := []string{
		"btp-iac-govern/SKILL.md",
		"btp-iac-scenario/SKILL.md",
		"btp-iac-accounts/SKILL.md",
		"btp-iac-services/SKILL.md",
		"btp-iac-security/SKILL.md",
		"btp-iac-connectivity/SKILL.md",
		"btp-iac-tasks/SKILL.md",
		"btp-iac-design/SKILL.md",
		"btp-iac-generate/SKILL.md",
	}
	for _, path := range linearFlowSkills {
		data, err := Commands.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := strings.Count(string(data), "## Next step"); got != 1 {
			t.Errorf("%s has %d Next step sections, want 1", path, got)
		}
	}

	for _, path := range []string{"btp-iac-analyse/SKILL.md", "btp-iac-next/SKILL.md"} {
		data, err := Commands.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "## Next step") {
			t.Errorf("%s must not have a linear-flow Next step footer", path)
		}
	}
}

func TestTasksSkillHasNoDuplicateNextStepInstruction(t *testing.T) {
	data, err := Commands.ReadFile("btp-iac-tasks/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "Next step: /btp-iac.design") {
		t.Error("tasks skill has a duplicate next-step instruction in its output template")
	}
}
