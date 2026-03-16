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
	)

	cmd := &cobra.Command{
		Use:   "tail",
		Short: "Stream logs from configured sources",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfgPath == "" {
				cfgPath = config.DefaultPath()
			}

			conns, err := buildConnectors(cfgPath, sources)
			if err != nil {
				return err
			}

			if len(conns) == 0 {
				fmt.Println("No sources configured. Run 'logmx init' then edit ~/.config/logmx/config.yaml")
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

	cmd.Flags().StringVar(&sources, "source", "", "Comma-separated source names to tail (default: all)")
	cmd.Flags().StringVar(&level, "level", "", "Filter by log level (info, warn, error, debug)")
	cmd.Flags().StringVar(&cfgPath, "config", "", "Path to config file")

	return cmd
}

func buildConnectors(cfgPath string, sourceFilter string) ([]connectors.Connector, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		// If no config exists, fall back to demo connectors
		if config.IsNotExist(err) {
			fmt.Println("No config found, using demo sources. Run 'logmx init' to configure.")
			return []connectors.Connector{
				&demo.DemoConnector{Source: "vercel"},
				&demo.DemoConnector{Source: "railway"},
				&demo.DemoConnector{Source: "gcp"},
			}, nil
		}
		return nil, err
	}

	// Load auth tokens
	store := auth.NewStore()
	if err := store.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load auth tokens: %v\n", err)
	}

	allowed := make(map[string]bool)
	if sourceFilter != "" {
		for _, s := range strings.Split(sourceFilter, ",") {
			allowed[strings.TrimSpace(s)] = true
		}
	}

	var conns []connectors.Connector
	for _, src := range cfg.Sources {
		if len(allowed) > 0 && !allowed[src.Name] {
			continue
		}

		c, err := connectorForSource(src, store)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping source %q: %v\n", src.Name, err)
			continue
		}
		conns = append(conns, c)
	}

	return conns, nil
}

func connectorForSource(src config.Source, store *auth.Store) (connectors.Connector, error) {
	switch src.Provider {
	case "demo":
		return &demo.DemoConnector{Source: src.Name}, nil

	case "vercel":
		token, err := store.Get("vercel")
		if err != nil {
			return nil, err
		}
		return &connVercel.Connector{
			Source:    src.Name,
			ProjectID: src.Project,
			Token:    token,
		}, nil

	case "railway":
		token, err := store.Get("railway")
		if err != nil {
			return nil, err
		}
		return &connRailway.Connector{
			Source:    src.Name,
			ServiceID: src.Service,
			Token:    token,
		}, nil

	default:
		return nil, fmt.Errorf("unknown provider %q", src.Provider)
	}
}
