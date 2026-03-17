package commands

import (
	"github.com/spf13/cobra"
)

func Root() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "logmx",
		Short:         "Aggregate and stream logs from multiple cloud platforms",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(setupCmd())
	cmd.AddCommand(authCmd())
	cmd.AddCommand(sourceCmd())
	cmd.AddCommand(tailCmd())

	return cmd
}
