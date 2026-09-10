// Command btp-iac is the spec-driven Terraform toolkit for SAP BTP.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/SAP/btp-iac-spec-kit/internal/agentselect"
	"github.com/SAP/btp-iac-spec-kit/internal/mcpcheck"
	"github.com/SAP/btp-iac-spec-kit/internal/preflight"
	"github.com/SAP/btp-iac-spec-kit/internal/scaffold"
	"github.com/SAP/btp-iac-spec-kit/internal/ui"
	"github.com/SAP/btp-iac-spec-kit/skills"
	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:          "btp-iac",
		Short:        "Spec-Driven Terraform Toolkit for SAP BTP",
		SilenceUsage: true,
	}
	root.AddCommand(initCmd())
	return root
}

// initIntent holds the resolved decision before any scaffold work runs.
type initIntent struct {
	dir  string
	mode scaffold.Mode
}

// validProjectName rejects names with path separators or leading dots to
// prevent directory traversal and hidden-directory confusion.
func validProjectName(name string) error {
	if strings.ContainsAny(name, `/\`) || strings.HasPrefix(name, ".") {
		return fmt.Errorf("invalid project name %q: must not contain path separators or start with '.'", name)
	}
	return nil
}

// resolveIntent determines what to do based on the optional name arg.
// When name is provided we skip the interactive menu.
// When name is absent we show the three-option menu.
func resolveIntent(name string) (initIntent, error) {
	if name != "" {
		if err := validProjectName(name); err != nil {
			return initIntent{}, err
		}
		if _, err := os.Stat(name); err == nil {
			// Named directory exists: pick mode based on whether it's already a
			// btp-iac project (has specs/ or memory/) or a bare repo to adopt.
			if scaffold.IsProject(name) {
				return initIntent{dir: name, mode: scaffold.ModeAgentOnly}, nil
			}
			return initIntent{dir: name, mode: scaffold.ModeAdopt}, nil
		}
		return initIntent{dir: name, mode: scaffold.ModeFresh}, nil
	}

	const (
		optFresh     = "Create a new project"
		optAdopt     = "Use existing directory (I already have Terraform files here)"
		optAgentOnly = "Add / update AI agent in current project"
	)

	var choice string
	if err := survey.AskOne(&survey.Select{
		Message: "What do you want to do?",
		Options: []string{optFresh, optAdopt, optAgentOnly},
	}, &choice); err != nil {
		return initIntent{}, fmt.Errorf("prompt: %w", err)
	}

	switch choice {
	case optFresh:
		var projectName string
		if err := survey.AskOne(
			&survey.Input{Message: "Project name:"},
			&projectName,
			survey.WithValidator(survey.Required),
			survey.WithValidator(func(val any) error {
				return validProjectName(fmt.Sprintf("%v", val))
			}),
		); err != nil {
			return initIntent{}, fmt.Errorf("prompt: %w", err)
		}
		return initIntent{dir: projectName, mode: scaffold.ModeFresh}, nil

	case optAdopt:
		cwd, err := os.Getwd()
		if err != nil {
			return initIntent{}, fmt.Errorf("getwd: %w", err)
		}
		return initIntent{dir: cwd, mode: scaffold.ModeAdopt}, nil

	default: // optAgentOnly
		cwd, err := os.Getwd()
		if err != nil {
			return initIntent{}, fmt.Errorf("getwd: %w", err)
		}
		return initIntent{dir: cwd, mode: scaffold.ModeAgentOnly}, nil
	}
}

func initCmd() *cobra.Command {
	var agentFlag string

	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Bootstrap or update a SAP BTP IaC project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if w := preflight.Warn("terraform"); w != "" {
				cmd.PrintErrln(ui.Warn(w))
			}

			var name string
			if len(args) == 1 {
				name = args[0]
			}

			// Fail early in non-TTY environments before showing any prompts.
			if err := agentselect.CheckTTYOrFlag(agentFlag); err != nil {
				return err
			}

			intent, err := resolveIntent(name)
			if err != nil {
				return err
			}

			agents, err := agentselect.Select(agentFlag)
			if err != nil {
				return err
			}

			agentIDs := make([]string, len(agents))
			for i, a := range agents {
				agentIDs[i] = a.ID
			}
			for _, w := range mcpcheck.Run(mcpcheck.DefaultChecks, agentIDs) {
				cmd.PrintErrln(ui.Warn(w))
			}

			if intent.mode == scaffold.ModeAdopt {
				var ok bool
				if err := survey.AskOne(&survey.Confirm{
					Message: fmt.Sprintf("This will create specs/, memory/, and terraform/ in %s. Continue?", filepath.Base(intent.dir)),
					Default: true,
				}, &ok); err != nil {
					return fmt.Errorf("prompt: %w", err)
				}
				if !ok {
					return nil
				}
				cmd.PrintErrln(ui.Warn("AI agents will generate Terraform files inside terraform/. Move your existing .tf files there if you want them alongside generated code."))
			}

			var warning string
			if intent.mode == scaffold.ModeFresh {
				warning, err = scaffold.Scaffold(intent.dir, skills.Commands, agents)
			} else {
				warning, err = scaffold.Apply(intent.dir, intent.mode, skills.Commands, agents)
			}
			if err != nil {
				return err
			}
			if warning != "" {
				cmd.PrintErrln(ui.Warn(warning))
			}

			ui.PrintSuccess(intent.dir, agentIDs, intent.mode, warning == "")
			return nil
		},
	}

	cmd.Flags().StringVar(&agentFlag, "agent", "", "comma-separated list of agents to configure (claude, cursor, copilot)")
	return cmd
}
