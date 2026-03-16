package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lucasnevespereira/logmx/internal/config"
)

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create a default config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := config.DefaultPath()
			if err := config.Init(path); err != nil {
				return err
			}
			fmt.Printf("Config created at %s\n", path)
			fmt.Println("Edit it to add your log sources.")
			return nil
		},
	}
}
