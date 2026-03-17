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
	"github.com/lucasnevespereira/logmx/internal/config"
	"github.com/lucasnevespereira/logmx/internal/log"
	"github.com/lucasnevespereira/logmx/internal/provider"
	"github.com/lucasnevespereira/logmx/internal/provider/demo"
	provRailway "github.com/lucasnevespereira/logmx/internal/provider/railway"
	provVercel "github.com/lucasnevespereira/logmx/internal/provider/vercel"
)

func tailCmd() *cobra.Command {
	var (
		sources string
		level   string
		cfgPath string
		limit   int
		follow  bool
	)

	cmd := &cobra.Command{
		Use:   "tail",
		Short: "Show recent logs from configured sources",
		Long:  "Fetch recent logs from all configured sources.\nUse -f to stream logs in real time.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfgPath == "" {
				cfgPath = config.DefaultPath()
			}

			conns, err := buildConnectors(cfgPath, sources, limit, follow)
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

			var levelFilter log.LogLevel
			if level != "" {
				levelFilter = log.LogLevel(strings.ToUpper(level))
			}

			for entry := range ch {
				if levelFilter != "" && entry.Level != levelFilter {
					continue
				}
				log.PrintEntry(entry)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&sources, "source", "s", "", "Comma-separated source names (default: all)")
	cmd.Flags().StringVar(&level, "level", "", "Filter by log level (info, warn, error, debug)")
	cmd.Flags().StringVar(&cfgPath, "config", "", "Path to config file")
	cmd.Flags().IntVarP(&limit, "limit", "n", 100, "Number of recent logs per source")
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Stream logs in real time")

	return cmd
}

func buildConnectors(cfgPath string, sourceFilter string, limit int, follow bool) ([]provider.Connector, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		if config.IsNotExist(err) {
			fmt.Println("No config found, using demo sources. Run 'logmx setup' to configure.")
			return []provider.Connector{
				&demo.DemoConnector{Source: "vercel", Follow: follow},
				&demo.DemoConnector{Source: "railway", Follow: follow},
			}, nil
		}
		return nil, err
	}

	allowed := make(map[string]bool)
	if sourceFilter != "" {
		for _, s := range strings.Split(sourceFilter, ",") {
			allowed[strings.TrimSpace(s)] = true
		}
	}

	// Load token store for providers that need it (e.g. Vercel).
	store, err := config.LoadAuth(config.DefaultAuthPath())
	if err != nil {
		return nil, fmt.Errorf("loading auth: %w", err)
	}

	var providerNames []string
	var conns []provider.Connector
	for _, src := range cfg.Sources {
		if len(allowed) > 0 && !allowed[src.Name] {
			continue
		}
		providerNames = append(providerNames, src.Provider)

		c, err := connectorForSource(src, store.Tokens[src.Provider], limit, follow)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping source %q: %v\n", src.Name, err)
			continue
		}
		conns = append(conns, c)
	}

	if missing := provider.MissingDeps(providerNames); len(missing) > 0 {
		for _, dep := range missing {
			fmt.Fprintf(os.Stderr, "missing: %s — install with: %s\n", dep.Name, dep.InstallCmd)
		}
		return nil, fmt.Errorf("install missing dependencies and try again")
	}

	return conns, nil
}

func connectorForSource(src config.Source, token string, limit int, follow bool) (provider.Connector, error) {
	switch src.Provider {
	case "demo":
		return &demo.DemoConnector{Source: src.Name, Follow: follow}, nil

	case "vercel":
		return &provVercel.Connector{
			Source:    src.Name,
			ProjectID: src.Project,
			Token:     token,
			Limit:     limit,
			Follow:    follow,
		}, nil

	case "railway":
		return &provRailway.Connector{
			Source:        src.Name,
			ProjectID:     src.Project,
			ServiceID:     src.Service,
			EnvironmentID: src.Environment,
			Limit:         limit,
			Follow:        follow,
		}, nil

	default:
		return nil, fmt.Errorf("unknown provider %q", src.Provider)
	}
}
