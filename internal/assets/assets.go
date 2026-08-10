// Package assets embeds the btp-iac agent command files into the binary.
package assets

import "embed"

// Commands holds the nine .md agent command files installed by btp-iac init.
//
//go:embed commands/*.md
var Commands embed.FS
