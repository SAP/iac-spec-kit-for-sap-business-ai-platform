// Command btp-iac is the spec-driven Terraform toolkit for SAP BTP.
package main

import (
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/SAP/btp-iac-spec-kit/internal/agentselect"
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

func initCmd() *cobra.Command {
	var agentFlag string

	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Bootstrap a new SAP BTP IaC project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if w := preflight.Warn("terraform"); w != "" {
				cmd.PrintErrln(ui.Warn(w))
			}

			agents, err := agentselect.Select(agentFlag)
			if err != nil {
				return err
			}

			var name string
			if len(args) == 1 {
				name = args[0]
			} else {
				if err := survey.AskOne(&survey.Input{Message: "Project name:"}, &name, survey.WithValidator(survey.Required)); err != nil {
					return err
				}
			}
			warning, err := scaffold.Scaffold(name, skills.Commands, agents)
			if err != nil {
				return err
			}
			if warning != "" {
				cmd.PrintErrln(ui.Warn(warning))
			}

			agentIDs := make([]string, len(agents))
			for i, a := range agents {
				agentIDs[i] = a.ID
			}

			ui.PrintSuccess(name, agentIDs, warning == "")
			return nil
		},
	}

	cmd.Flags().StringVar(&agentFlag, "agent", "", "comma-separated list of agents to configure (claude, cursor, copilot)")
	return cmd
}
