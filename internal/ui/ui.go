// Package ui renders terminal output for the btp-iac CLI.
package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
)

const (
	boxWidth    = 80
	cmdColWidth = 22
)

var (
	successMark = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(10)).Bold(true).Render("✓")
	checkLine   = lipgloss.NewStyle().PaddingLeft(2)

	// Left-border panel: clean single-line accent instead of a full box.
	panel = lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.ANSIColor(12)).
		PaddingLeft(2).
		PaddingRight(2)

	headingStyle = lipgloss.NewStyle().Bold(true)
	sectionTitle = lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(12)).
			Bold(true)

	cmdStyle  = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(14)).Bold(true)
	stepStyle = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(11))
	warnStyle = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(11)).Background(lipgloss.ANSIColor(0)).Bold(true)
	warnMsg   = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(15)).Background(lipgloss.ANSIColor(0)).Bold(true)
)

// Warn renders msg as a styled warning line for use on stderr.
func Warn(msg string) string {
	return warnStyle.Render("Warning:") + " " + warnMsg.Render(msg)
}

// PrintSuccess prints the post-init success message and workflow guide.
// gitInitialised should be true when git init ran successfully.
func PrintSuccess(name string, agentIDs []string, gitInitialised bool) {
	printSuccess(os.Stdout, name, agentIDs, gitInitialised)
}

func printSuccess(w io.Writer, name string, agentIDs []string, gitInitialised bool) {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(checkLine.Render(successMark + "  Project " + fmt.Sprintf("%q", name) + " created"))
	sb.WriteString("\n")
	sb.WriteString(checkLine.Render(successMark + "  " + strings.Join(agentIDs, ", ") + " configured"))
	sb.WriteString("\n")
	if gitInitialised {
		sb.WriteString(checkLine.Render(successMark + "  Git repository initialised"))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	var body strings.Builder

	body.WriteString(headingStyle.Render("Getting started"))
	body.WriteString("\n\n")
	body.WriteString("I. Open the project in your terminal\n\n")
	fmt.Fprintf(&body, "   %s\n\n", cmdStyle.Render("$ cd "+name))
	body.WriteString("II. Open your AI agent and run these in order\n\n")

	section(&body, "Define your guardrails", []stepRow{
		{"-", "btp-iac.govern", "Set guardrails — regions, naming, cost policies. Skip to use defaults (optional)."},
	})

	section(&body, "Define your infrastructure", []stepRow{
		{"1)", "btp-iac.scenario", "Describe your app."},
		{"2)", "btp-iac.analyse", "Scan source code to extract service dependencies automatically (optional)."},
		{"3)", "btp-iac.accounts", "Map your app to BTP directories and subaccounts."},
		{"4)", "btp-iac.services", "Resolve which BTP services each subaccount needs."},
		{"5)", "btp-iac.security", "Set up IdP trust, role collections, destinations."},
	})

	section(&body, "Generate Terraform", []stepRow{
		{"6)", "btp-iac.tasks", "Build a dependency-ordered execution plan."},
		{"7)", "btp-iac.design", "Plan the Terraform file and module layout."},
		{"8)", "btp-iac.generate", "Write and validate all Terraform HCL."},
	})

	sb.WriteString(panel.Render(body.String()))
	sb.WriteString("\n\n")

	_, _ = fmt.Fprint(w, sb.String())
}

type stepRow struct{ step, cmd, desc string }

func section(b *strings.Builder, title string, rows []stepRow) {
	fmt.Fprintf(b, "%s\n\n", sectionTitle.Render(title))
	for _, r := range rows {
		renderedCmd := cmdStyle.Render(r.cmd)
		// Pad based on visible width so ANSI escapes don't throw off alignment.
		pad := cmdColWidth - lipgloss.Width(renderedCmd)
		pad = max(pad, 1)
		fmt.Fprintf(b, "%s  %s%s%s\n",
			stepStyle.Render(r.step),
			renderedCmd,
			strings.Repeat(" ", pad),
			r.desc,
		)
	}
	b.WriteString("\n")
}
