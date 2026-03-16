package commands

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lucasnevespereira/logmx/internal/auth"
	"github.com/lucasnevespereira/logmx/internal/config"
	"github.com/lucasnevespereira/logmx/internal/provider/railway"
	"github.com/lucasnevespereira/logmx/internal/provider/vercel"
)

type pickableSource struct {
	Name     string
	Provider string
	Project  string
	Service  string
}

func addCmd() *cobra.Command {
	var cfgPath string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Interactively add log sources from your providers",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfgPath == "" {
				cfgPath = config.DefaultPath()
			}

			store := auth.NewStore()
			if err := store.Load(); err != nil {
				return err
			}

			providers := store.Providers()
			if len(providers) == 0 {
				fmt.Println("No providers authenticated.")
				fmt.Println("Run 'logmx auth <provider>' first. Supported: vercel, railway")
				return nil
			}

			// Pick provider
			fmt.Println("Authenticated providers:")
			for i, p := range providers {
				fmt.Printf("  [%d] %s\n", i+1, p)
			}
			fmt.Println()

			provider, err := promptChoice("Provider", len(providers))
			if err != nil {
				return err
			}
			selectedProvider := providers[provider]

			token, err := store.Get(selectedProvider)
			if err != nil {
				return err
			}

			// Fetch sources from provider
			fmt.Printf("Fetching projects from %s...\n", selectedProvider)
			sources, err := fetchSources(selectedProvider, token)
			if err != nil {
				return fmt.Errorf("fetching projects: %w", err)
			}

			if len(sources) == 0 {
				fmt.Println("No projects found.")
				return nil
			}

			fmt.Println()
			for i, s := range sources {
				label := s.Name
				if s.Service != "" {
					label = fmt.Sprintf("%s / %s", s.Project, s.Name)
				}
				fmt.Printf("  [%d] %s\n", i+1, label)
			}
			fmt.Println()

			picks, err := promptMultiChoice("Select sources (comma-separated, e.g. 1,3)", len(sources))
			if err != nil {
				return err
			}

			// Load or init config
			cfg, err := config.Load(cfgPath)
			if err != nil {
				if config.IsNotExist(err) {
					if err := config.Init(cfgPath); err != nil {
						return err
					}
					cfg = &config.Config{}
				} else {
					return err
				}
			}

			// Deduplicate: check existing source names
			existing := make(map[string]bool)
			for _, s := range cfg.Sources {
				existing[s.Name] = true
			}

			added := 0
			for _, idx := range picks {
				s := sources[idx]

				name := s.Name
				if existing[name] {
					fmt.Printf("  skipping %q (already configured)\n", name)
					continue
				}

				src := config.Source{
					Name:     name,
					Provider: selectedProvider,
					Project:  s.Project,
					Service:  s.Service,
				}
				cfg.Sources = append(cfg.Sources, src)
				existing[name] = true
				added++
			}

			if err := config.Save(cfgPath, cfg); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			fmt.Printf("\nAdded %d source(s). Run 'logmx tail' to start streaming.\n", added)
			return nil
		},
	}

	cmd.Flags().StringVar(&cfgPath, "config", "", "Path to config file")
	return cmd
}

func fetchSources(provider, token string) ([]pickableSource, error) {
	switch provider {
	case "vercel":
		return fetchVercelSources(token)
	case "railway":
		return fetchRailwaySources(token)
	default:
		return nil, fmt.Errorf("unknown provider %q", provider)
	}
}

func fetchVercelSources(token string) ([]pickableSource, error) {
	c := vercel.NewClient(token)
	projects, err := c.ListProjects()
	if err != nil {
		return nil, err
	}

	var sources []pickableSource
	for _, p := range projects {
		sources = append(sources, pickableSource{
			Name:    p.Name,
			Provider: "vercel",
			Project: p.ID,
		})
	}
	return sources, nil
}

func fetchRailwaySources(token string) ([]pickableSource, error) {
	c := railway.NewClient(token)
	projects, err := c.ListProjects()
	if err != nil {
		return nil, err
	}

	var sources []pickableSource
	for _, p := range projects {
		for _, svc := range p.Services {
			sources = append(sources, pickableSource{
				Name:    svc.Name,
				Provider: "railway",
				Project: p.ID,
				Service: svc.ID,
			})
		}
	}
	return sources, nil
}

func promptChoice(label string, max int) (int, error) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("%s [1-%d]: ", label, max)
	if !scanner.Scan() {
		return 0, fmt.Errorf("no input")
	}

	n, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil || n < 1 || n > max {
		return 0, fmt.Errorf("invalid choice — enter a number between 1 and %d", max)
	}
	return n - 1, nil
}

func promptMultiChoice(label string, max int) ([]int, error) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("%s: ", label)
	if !scanner.Scan() {
		return nil, fmt.Errorf("no input")
	}

	var indices []int
	for _, part := range strings.Split(scanner.Text(), ",") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || n < 1 || n > max {
			return nil, fmt.Errorf("invalid choice %q — enter numbers between 1 and %d", part, max)
		}
		indices = append(indices, n-1)
	}

	if len(indices) == 0 {
		return nil, fmt.Errorf("no sources selected")
	}
	return indices, nil
}
