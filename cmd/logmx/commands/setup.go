package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/lucasnevespereira/logmx/internal/config"
	"github.com/lucasnevespereira/logmx/internal/provider"
)

var (
	title    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	success  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	warning  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	dim      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

func setupCmd() *cobra.Command {
	var cfgPath string

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Interactive wizard to configure logmx",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfgPath == "" {
				cfgPath = config.DefaultPath()
			}
			return runSetup(cfgPath)
		},
	}

	cmd.Flags().StringVar(&cfgPath, "config", "", "Path to config file")
	return cmd
}

func runSetup(cfgPath string) error {
	fmt.Println(title.Render("\n  logmx setup\n"))

	// Step 1: Pick providers
	fmt.Println(title.Render("  Providers"))
	fmt.Println()

	var chosen []string
	options := make([]huh.Option[string], len(supportedProviders))
	for i, p := range supportedProviders {
		options[i] = huh.NewOption(p, p)
	}

	err := huh.NewMultiSelect[string]().
		Title("Which providers do you use?").
		Options(options...).
		Value(&chosen).
		Run()
	if err != nil {
		return err
	}
	if len(chosen) == 0 {
		fmt.Println(warning.Render("  No providers selected."))
		return nil
	}

	// Step 2: Authenticate (paste tokens)
	fmt.Println()
	fmt.Println(title.Render("  Authentication"))
	fmt.Println()

	store, err := config.LoadAuth(config.DefaultAuthPath())
	if err != nil {
		return err
	}

	var authenticated []string
	for _, prov := range chosen {
		if token := store.Tokens[prov]; token != "" {
			name, err := validateToken(prov, token)
			if err == nil {
				fmt.Printf("  %s %s %s\n", success.Render("✓"), prov, dim.Render(fmt.Sprintf("(%s)", name)))
				authenticated = append(authenticated, prov)
				continue
			}
			fmt.Printf("  %s %s %s\n", warning.Render("✗"), prov, dim.Render("(token expired)"))
		} else {
			fmt.Printf("  %s %s %s\n", warning.Render("✗"), prov, dim.Render("(no token)"))
		}

		if err := runAuth(prov); err != nil {
			fmt.Printf("  %s\n", errStyle.Render(fmt.Sprintf("Skipped: %v", err)))
			continue
		}

		// Reload store after auth saved the token
		store, err = config.LoadAuth(config.DefaultAuthPath())
		if err != nil {
			return err
		}
		authenticated = append(authenticated, prov)
	}

	if len(authenticated) == 0 {
		fmt.Println(errStyle.Render("\n  No providers authenticated. Run 'logmx auth <provider>' to add a token."))
		return nil
	}

	// Step 3: Check CLI dependencies (needed for log streaming)
	fmt.Println()
	fmt.Println(title.Render("  Streaming dependencies"))
	fmt.Println()

	installCLIs(authenticated)

	// Step 4: Pick sources via API
	fmt.Println()
	fmt.Println(title.Render("  Sources"))
	fmt.Println()

	var allSources []config.Source
	for _, prov := range authenticated {
		token := store.Tokens[prov]
		dep := provider.ProviderDeps[prov]

		fmt.Printf("  Fetching projects from %s...\n", dep.Name)
		opts, err := fetchProjectOptions(prov, token)
		if err != nil {
			fmt.Printf("  %s\n", errStyle.Render(fmt.Sprintf("Failed: %v", err)))
			continue
		}
		if len(opts) == 0 {
			fmt.Printf("  %s\n", dim.Render("No projects found."))
			continue
		}

		var labels []string
		for _, o := range opts {
			label := o.Name
			if o.Service != "" {
				label = fmt.Sprintf("%s / %s", o.Project, o.Name)
			}
			labels = append(labels, label)
		}

		var selected []int
		huhOptions := make([]huh.Option[int], len(labels))
		for i, label := range labels {
			huhOptions[i] = huh.NewOption(label, i)
		}

		err = huh.NewMultiSelect[int]().
			Title(fmt.Sprintf("Select %s sources", dep.Name)).
			Options(huhOptions...).
			Value(&selected).
			Run()
		if err != nil {
			return err
		}

		for _, idx := range selected {
			o := opts[idx]
			allSources = append(allSources, config.Source{
				Name:     o.Name,
				Provider: prov,
				Project:  o.Project,
				Service:  o.Service,
			})
		}
	}

	if len(allSources) == 0 {
		fmt.Println(warning.Render("  No sources selected."))
		return nil
	}

	// Step 5: Save config
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
	for _, s := range allSources {
		if existing[s.Name] {
			continue
		}
		cfg.Sources = append(cfg.Sources, s)
		existing[s.Name] = true
		added++
	}

	if err := config.Save(cfgPath, cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Println()
	fmt.Println(success.Render(fmt.Sprintf("  Added %d source(s) to %s", added, cfgPath)))
	fmt.Println(dim.Render("  Run 'logmx tail' to start streaming.\n"))
	return nil
}

func installCLIs(providers []string) {
	for _, prov := range providers {
		dep := provider.ProviderDeps[prov]

		if _, err := exec.LookPath(dep.Binary); err == nil {
			fmt.Printf("  %s %s\n", success.Render("✓"), dep.Name)
			continue
		}

		fmt.Printf("  %s %s %s\n", warning.Render("✗"), dep.Name, dim.Render("(needed for streaming)"))

		var install bool
		err := huh.NewConfirm().
			Title(fmt.Sprintf("Install %s? (%s)", dep.Name, dep.InstallCmd)).
			Value(&install).
			Run()
		if err != nil || !install {
			continue
		}

		fmt.Printf("  Installing %s...\n", dep.Name)
		parts := strings.Fields(dep.InstallCmd)
		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("  %s\n", errStyle.Render(fmt.Sprintf("Failed: %v", err)))
			continue
		}

		fmt.Printf("  %s %s installed\n", success.Render("✓"), dep.Name)
	}
}
