// Package preflight checks that required host binaries are available before
// sap-iac init creates any files on disk.
package preflight

import (
	"fmt"
	"os/exec"
)

// Available reports whether name can be resolved on PATH. It is suitable for
// optional capabilities, unlike Check which is intended for required tools.
func Available(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// Check returns an error if the named binary is not found on $PATH.
func Check(name string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("%q is required but was not found on $PATH — please install it and try again", name)
	}
	return nil
}

// Warn returns a warning message if the named binary is not found on $PATH,
// or an empty string if it is found.
func Warn(name string) string {
	if !Available(name) {
		return fmt.Sprintf("%q was not found on $PATH — install it before running terraform commands.", name)
	}
	return ""
}
