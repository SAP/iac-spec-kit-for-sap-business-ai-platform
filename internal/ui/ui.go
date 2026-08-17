// Package ui renders terminal output for the btp-iac CLI.
package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

// PrintSuccess prints the post-init success message and workflow guide.
func PrintSuccess(name string, agentIDs []string) {
	printSuccess(os.Stdout, name, agentIDs)
}

func printSuccess(w io.Writer, name string, agentIDs []string) {
	const inner = 78 // visible character width between │ and │

	// pad pads s to inner visible characters.
	pad := func(s string) string {
		n := utf8.RuneCountInString(s)
		if n >= inner {
			return s
		}
		return s + strings.Repeat(" ", inner-n)
	}

	l := func(s string) { fmt.Fprintf(w, "│ %s │\n", pad(s)) }
	blank := func() { l("") }

	divider := func(label string) {
		labelWidth := utf8.RuneCountInString(label)
		dashes := strings.Repeat("─", inner-labelWidth-4)
		fmt.Fprintf(w, "├──  %s  %s┤\n", label, dashes)
	}

	row := func(num, command, desc string) {
		l(fmt.Sprintf("  %s  %-22s%s", num, command, desc))
	}
	row2 := func(desc string) {
		l(fmt.Sprintf("  %s  %-22s%s", " ", "", desc))
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "  ✓  Project %q created\n", name)
	fmt.Fprintf(w, "  ✓  %s configured\n", strings.Join(agentIDs, ", "))
	fmt.Fprintf(w, "  ✓  Git repository initialised\n")
	fmt.Fprintln(w)

	fmt.Fprintf(w, "╭%s╮\n", strings.Repeat("─", inner+2))
	blank()
	l("  Getting started")
	blank()
	l("  1. Open the project in your terminal")
	blank()
	l(fmt.Sprintf("     $ cd %s", name))
	blank()
	l("  2. Open your AI agent and run these in order")
	blank()

	divider("Before you start")
	blank()
	row("○", "btp-iac.govern", "(optional) Set guardrails — regions,")
	row2("naming, cost policies. Skip to use defaults.")
	blank()

	divider("Define your infrastructure")
	blank()
	row("1", "btp-iac.scenario", "Describe your app.")
	row("2", "btp-iac.analyse", "(optional) Scan source code to extract")
	row2("service dependencies automatically.")
	row("3", "btp-iac.accounts", "Map your app to BTP directories and subaccounts.")
	row("4", "btp-iac.services", "Resolve which BTP services each subaccount needs.")
	row("5", "btp-iac.security", "Set up IdP trust, role collections, destinations.")
	blank()

	divider("Generate Terraform")
	blank()
	row("6", "btp-iac.tasks", "Build a dependency-ordered execution plan.")
	row("7", "btp-iac.design", "Plan the Terraform file and module layout.")
	row("8", "btp-iac.generate", "Write and validate all Terraform HCL.")
	blank()

	fmt.Fprintf(w, "╰%s╯\n", strings.Repeat("─", inner+2))
	fmt.Fprintln(w)
}
