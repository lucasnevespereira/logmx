package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/lucasnevespereira/logmx/internal/aggregator"
	"github.com/lucasnevespereira/logmx/internal/auth"
	"github.com/lucasnevespereira/logmx/internal/cli"
	"github.com/lucasnevespereira/logmx/internal/config"
	"github.com/lucasnevespereira/logmx/internal/connectors"
	"github.com/lucasnevespereira/logmx/internal/connectors/demo"
	connRailway "github.com/lucasnevespereira/logmx/internal/connectors/railway"
	connVercel "github.com/lucasnevespereira/logmx/internal/connectors/vercel"
	"github.com/lucasnevespereira/logmx/internal/models"
	"github.com/lucasnevespereira/logmx/internal/printer"
)

func tailCmd() *cobra.Command {
	var (
		sources string
		level   string
		cfgPath string
		limit   int
	)

	cmd := &cobra.Command{
		Use:   "tail",
		Short: "Show recent logs from configured sources",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfgPath == "" {
				cfgPath = config.DefaultPath()
			}

			conns, err := buildConnectors(cfgPath, sources, limit)
			if err != nil {
				return err
			}

			if len(conns) == 0 {
				fmt.Println("No sources configured. Run 'logmx setup' to get started.")
				return nil
			}

			ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			agg := aggregator.New(conns)
			ch := agg.Run(ctx)

			var levelFilter models.LogLevel
			if level != "" {
				levelFilter = models.LogLevel(strings.ToUpper(level))
			}

			for entry := range ch {
				if levelFilter != "" && entry.Level != levelFilter {
					continue
				}
				printer.PrintEntry(entry)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&sources, "source", "", "Comma-separated source names (default: all)")
	cmd.Flags().StringVar(&level, "level", "", "Filter by log level (info, warn, error, debug)")
	cmd.Flags().StringVar(&cfgPath, "config", "", "Path to config file")
	cmd.Flags().IntVarP(&limit, "limit", "n", 100, "Number of recent logs per source")

	return cmd
}

func buildConnectors(cfgPath string, sourceFilter string, limit int) ([]connectors.Connector, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		if config.IsNotExist(err) {
			fmt.Println("No config found, using demo sources. Run 'logmx setup' to configure.")
			return []connectors.Connector{
				&demo.DemoConnector{Source: "vercel"},
				&demo.DemoConnector{Source: "railway"},
			}, nil
		}
		return nil, err
	}

	store, err := auth.Load(auth.DefaultPath())
	if err != nil {
		return nil, fmt.Errorf("loading auth: %w", err)
	}

	allowed := make(map[string]bool)
	if sourceFilter != "" {
		for _, s := range strings.Split(sourceFilter, ",") {
			allowed[strings.TrimSpace(s)] = true
		}
	}

	var providers []string
	var conns []connectors.Connector
	for _, src := range cfg.Sources {
		if len(allowed) > 0 && !allowed[src.Name] {
			continue
		}
		providers = append(providers, src.Provider)

		token := store.Tokens[src.Provider]
		c, err := connectorForSource(src, token, limit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping source %q: %v\n", src.Name, err)
			continue
		}
		conns = append(conns, c)
	}

	if missing := cli.MissingDeps(providers); len(missing) > 0 {
		for _, dep := range missing {
			fmt.Fprintf(os.Stderr, "missing: %s — install with: %s\n", dep.Name, dep.InstallCmd)
		}
		return nil, fmt.Errorf("install missing dependencies and try again")
	}

	return conns, nil
}

func connectorForSource(src config.Source, token string, limit int) (connectors.Connector, error) {
	switch src.Provider {
	case "demo":
		return &demo.DemoConnector{Source: src.Name}, nil

	case "vercel":
		return &connVercel.Connector{
			Source:    src.Name,
			ProjectID: src.Project,
			Token:     token,
			Limit:     limit,
		}, nil

	case "railway":
		return &connRailway.Connector{
			Source:    src.Name,
			ProjectID: src.Project,
			ServiceID: src.Service,
			Token:     token,
			Limit:     limit,
		}, nil

	default:
		return nil, fmt.Errorf("unknown provider %q", src.Provider)
	}
}
