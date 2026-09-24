// Package skills embeds the btp-iac agent skill files into the binary.
package skills

import "embed"

// Commands holds the btp-iac agent skill files installed by btp-iac init.
//
//go:embed */SKILL.md
//go:embed */service-params-catalogue.yaml
var Commands embed.FS
