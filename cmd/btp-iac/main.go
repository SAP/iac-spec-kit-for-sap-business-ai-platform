// Command btp-iac is the spec-driven Terraform toolkit for SAP BTP.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/SAP/btp-iac-sdd/internal/agentselect"
	"github.com/SAP/btp-iac-sdd/internal/assets"
	"github.com/SAP/btp-iac-sdd/internal/preflight"
	"github.com/SAP/btp-iac-sdd/internal/scaffold"
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
		Short:        "Spec-driven Terraform toolkit for SAP BTP",
		SilenceUsage: true,
	}
	root.AddCommand(initCmd())
	return root
}

func initCmd() *cobra.Command {
	var agentFlag string

	cmd := &cobra.Command{
		Use:   "init <name>",
		Short: "Bootstrap a new BTP IaC project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, dep := range []string{"terraform", "git"} {
				if err := preflight.Check(dep); err != nil {
					return err
				}
			}

			agents, err := agentselect.Select(agentFlag)
			if err != nil {
				return err
			}

			name := args[0]
			if err := scaffold.Scaffold(name, assets.Commands, agents); err != nil {
				return err
			}

			agentIDs := make([]string, len(agents))
			for i, a := range agents {
				agentIDs[i] = a.ID
			}

			fmt.Printf("\nProject %q is ready (agents: %s).\n\n", name, strings.Join(agentIDs, ", "))
			fmt.Printf("Next steps:\n")
			fmt.Printf("  1. Open the project in your AI tool:  cd %s\n", name)
			fmt.Printf("  2. Describe your app:                 /btp-iac:scenario\n")
			fmt.Printf("  3. (Optional) Set guardrails first:   /btp-iac:govern\n")
			return nil
		},
	}

	cmd.Flags().StringVar(&agentFlag, "agent", "", "comma-separated list of agents to configure (claude, cursor, copilot)")
	return cmd
}
