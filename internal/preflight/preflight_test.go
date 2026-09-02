package preflight

import "testing"

func TestWarn(t *testing.T) {
	if got := Warn("sh"); got != "" {
		t.Errorf("Warn(found binary) = %q, want empty", got)
	}
	if got := Warn("btp-iac-nonexistent-binary-xyz"); got == "" {
		t.Error("Warn(missing binary) = empty, want non-empty warning")
	}
}

func TestCheck(t *testing.T) {
	tests := []struct {
		name    string
		binary  string
		wantErr bool
	}{
		{"found", "sh", false},
		{"not found", "btp-iac-nonexistent-binary-xyz", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Check(tt.binary)
			if (err != nil) != tt.wantErr {
				t.Errorf("Check(%q) error = %v, wantErr %v", tt.binary, err, tt.wantErr)
			}
		})
	}
}
