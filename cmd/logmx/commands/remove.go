package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lucasnevespereira/logmx/internal/config"
)

func removeCmd() *cobra.Command {
	var cfgPath string

	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a log source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfgPath == "" {
				cfgPath = config.DefaultPath()
			}

			name := args[0]

			cfg, err := config.Load(cfgPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			found := false
			filtered := cfg.Sources[:0]
			for _, s := range cfg.Sources {
				if s.Name == name {
					found = true
					continue
				}
				filtered = append(filtered, s)
			}

			if !found {
				return fmt.Errorf("source %q not found", name)
			}

			cfg.Sources = filtered
			if err := config.Save(cfgPath, cfg); err != nil {
				return err
			}

			fmt.Printf("Removed source %q.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&cfgPath, "config", "", "Path to config file")
	return cmd
}
