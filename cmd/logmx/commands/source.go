package commands

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/lucasnevespereira/logmx/internal/config"
	"github.com/lucasnevespereira/logmx/internal/provider/railway"
	"github.com/lucasnevespereira/logmx/internal/provider/vercel"
)

var supportedProviders = []string{"vercel", "railway"}

func sourceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "source",
		Short: "Manage log sources",
	}

	cmd.AddCommand(sourceAddCmd())
	cmd.AddCommand(sourceListCmd())
	cmd.AddCommand(sourceRemoveCmd())

	return cmd
}

// --- source add ---

type sourceOption struct {
	Name    string
	Project string
	Service string
}

func sourceAddCmd() *cobra.Command {
	var (
		cfgPath string
		from    string
	)

	cmd := &cobra.Command{
		Use:     "add",
		Short:   "Add log sources from a provider",
		Example: "  logmx source add\n  logmx source add --from vercel",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfgPath == "" {
				cfgPath = config.DefaultPath()
			}

			prov := strings.ToLower(from)

			if prov == "" {
				options := make([]huh.Option[string], len(supportedProviders))
				for i, p := range supportedProviders {
					options[i] = huh.NewOption(p, p)
				}

				err := huh.NewSelect[string]().
					Title("Pick a provider").
					Options(options...).
					Value(&prov).
					Run()
				if err != nil {
					return err
				}
			}

			if !isSupportedProvider(prov) {
				return fmt.Errorf("unknown provider %q — supported: %s",
					prov, strings.Join(supportedProviders, ", "))
			}

			store, err := config.LoadAuth(config.DefaultAuthPath())
			if err != nil {
				return err
			}

			token := store.Tokens[prov]
			if token == "" {
				return fmt.Errorf("no token for %s — run 'logmx auth %s' first", prov, prov)
			}

			fmt.Printf("Fetching projects from %s...\n", prov)
			sources, err := fetchProjectOptions(prov, token)
			if err != nil {
				return fmt.Errorf("fetching projects: %w", err)
			}

			if len(sources) == 0 {
				fmt.Println("No projects found.")
				return nil
			}

			var labels []string
			for _, s := range sources {
				label := s.Name
				if s.Service != "" {
					label = fmt.Sprintf("%s / %s", s.Project, s.Name)
				}
				labels = append(labels, label)
			}

			var selected []int
			huhOptions := make([]huh.Option[int], len(labels))
			for i, label := range labels {
				huhOptions[i] = huh.NewOption(label, i)
			}

			err = huh.NewMultiSelect[int]().
				Title("Select sources to add").
				Options(huhOptions...).
				Value(&selected).
				Run()
			if err != nil {
				return err
			}

			if len(selected) == 0 {
				fmt.Println("No sources selected.")
				return nil
			}

			cfg, err := config.Load(cfgPath)
			if err != nil {
				if config.IsNotExist(err) {
					cfg = &config.Config{}
				} else {
					return err
				}
			}

			existing := make(map[string]bool)
			for _, s := range cfg.Sources {
				existing[s.Name] = true
			}

			added := 0
			for _, idx := range selected {
				s := sources[idx]
				if existing[s.Name] {
					fmt.Printf("  skipping %q (already configured)\n", s.Name)
					continue
				}

				cfg.Sources = append(cfg.Sources, config.Source{
					Name:     s.Name,
					Provider: prov,
					Project:  s.Project,
					Service:  s.Service,
				})
				existing[s.Name] = true
				added++
			}

			if err := config.Save(cfgPath, cfg); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			fmt.Printf("Added %d source(s). Run 'logmx tail' to start streaming.\n", added)
			return nil
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "Provider to add sources from (vercel, railway)")
	cmd.Flags().StringVar(&cfgPath, "config", "", "Path to config file")
	return cmd
}

// --- source list ---

func sourceListCmd() *cobra.Command {
	var cfgPath string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List configured sources",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfgPath == "" {
				cfgPath = config.DefaultPath()
			}

			cfg, err := config.Load(cfgPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			if len(cfg.Sources) == 0 {
				fmt.Println("No sources configured. Run 'logmx source add' to add some.")
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

// --- source remove ---

func sourceRemoveCmd() *cobra.Command {
	var cfgPath string

	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a source",
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

// --- shared helpers ---

func isSupportedProvider(p string) bool {
	for _, s := range supportedProviders {
		if s == p {
			return true
		}
	}
	return false
}

func fetchProjectOptions(prov, token string) ([]sourceOption, error) {
	switch prov {
	case "vercel":
		c := vercel.NewClient(token)
		projects, err := c.ListProjects()
		if err != nil {
			return nil, err
		}
		var opts []sourceOption
		for _, p := range projects {
			opts = append(opts, sourceOption{Name: p.Name, Project: p.Name})
		}
		return opts, nil

	case "railway":
		c := railway.NewClient(token)
		projects, err := c.ListProjects()
		if err != nil {
			return nil, err
		}
		var opts []sourceOption
		for _, p := range projects {
			if len(p.Services) == 0 {
				opts = append(opts, sourceOption{Name: p.Name, Project: p.ID})
				continue
			}
			for _, svc := range p.Services {
				opts = append(opts, sourceOption{
					Name:    svc.Name,
					Project: p.ID,
					Service: svc.ID,
				})
			}
		}
		return opts, nil
	}
	return nil, fmt.Errorf("unknown provider %q", prov)
}
