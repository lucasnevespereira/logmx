package commands

import (
	"github.com/spf13/cobra"
)

func Root() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logmx",
		Short: "Aggregate and stream logs from multiple cloud platforms",
	}

	cmd.AddCommand(tailCmd())
	cmd.AddCommand(sourcesCmd())
	cmd.AddCommand(initCmd())
	cmd.AddCommand(authCmd())
	cmd.AddCommand(addCmd())
	cmd.AddCommand(removeCmd())

	return cmd
}
