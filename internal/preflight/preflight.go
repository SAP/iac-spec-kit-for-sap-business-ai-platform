// Package preflight checks that required host binaries are available before
// btp-iac init creates any files on disk.
package preflight

import (
	"fmt"
	"os/exec"
)

// Check returns an error if the named binary is not found on $PATH.
func Check(name string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("%q is required but was not found on $PATH — please install it and try again", name)
	}
	return nil
}
