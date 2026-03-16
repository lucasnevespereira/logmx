package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lucasnevespereira/logmx/internal/config"
)

func sourcesCmd() *cobra.Command {
	var cfgPath string

	cmd := &cobra.Command{
		Use:   "sources",
		Short: "List configured log sources",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfgPath == "" {
				cfgPath = config.DefaultPath()
			}

			cfg, err := config.Load(cfgPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			if len(cfg.Sources) == 0 {
				fmt.Println("No sources configured. Edit ~/.config/logmx/config.yaml to add sources.")
				return nil
			}

			fmt.Printf("%-15s %-10s %s\n", "NAME", "PROVIDER", "TARGET")
			for _, s := range cfg.Sources {
				target := s.Project
				if target == "" {
					target = s.Service
				}
				fmt.Printf("%-15s %-10s %s\n", s.Name, s.Provider, target)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&cfgPath, "config", "", "Path to config file")
	return cmd
}
