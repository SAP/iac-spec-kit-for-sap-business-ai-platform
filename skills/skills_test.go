package skills

import (
	"io/fs"
	"strings"
	"testing"
)

func TestAllSkillsContainPlatformValidationContract(t *testing.T) {
	err := fs.WalkDir(Commands, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, "/SKILL.md") {
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
