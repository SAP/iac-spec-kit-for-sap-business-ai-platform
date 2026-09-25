// Package skills embeds the sap-iac agent skill files into the binary.
package skills

import "embed"

// Commands holds the sap-iac agent skill files installed by sap-iac init.
//
//go:embed */SKILL.md
//go:embed */service-params-catalogue.yaml
var Commands embed.FS
