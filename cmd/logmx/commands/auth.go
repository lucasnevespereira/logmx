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

func authCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth <provider>",
		Short: "Authenticate with a cloud provider",
		Long: fmt.Sprintf("Authenticate with a cloud provider.\n\nSupported: %s",
			strings.Join(supportedProviders, ", ")),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prov := strings.ToLower(args[0])
			if !isSupportedProvider(prov) {
				return fmt.Errorf("unknown provider %q — supported: %s",
					prov, strings.Join(supportedProviders, ", "))
			}
			return runAuth(prov)
		},
	}
	return cmd
}

func runAuth(prov string) error {
	switch prov {
	case "railway":
		fmt.Println("Logging in to Railway CLI...")
		if err := railway.LoginBrowserless(); err != nil {
			return fmt.Errorf("railway login failed: %w", err)
		}
		fmt.Println("Railway authenticated.")
		return nil

	default:
		return runTokenAuth(prov)
	}
}

func runTokenAuth(prov string) error {
	urls := map[string]string{
		"vercel": "https://vercel.com/account/tokens",
	}
	fmt.Printf("Create a token at: %s\n\n", urls[prov])

	var token string
	err := huh.NewInput().
		Title(fmt.Sprintf("Paste your %s token", prov)).
		EchoMode(huh.EchoModePassword).
		Value(&token).
		Run()
	if err != nil {
		return err
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	fmt.Print("Verifying... ")
	name, err := validateToken(prov, token)
	if err != nil {
		fmt.Println("failed")
		return fmt.Errorf("invalid token: %w", err)
	}
	fmt.Printf("authenticated as %s\n", name)

	store, err := config.LoadAuth(config.DefaultAuthPath())
	if err != nil {
		return err
	}
	store.Tokens[prov] = token
	if err := config.SaveAuth(config.DefaultAuthPath(), store); err != nil {
		return fmt.Errorf("saving token: %w", err)
	}

	fmt.Printf("\nRun 'logmx setup' or 'logmx source add' to pick projects.\n")
	return nil
}

func validateToken(prov, token string) (string, error) {
	switch prov {
	case "vercel":
		c := vercel.NewClient(token)
		u, err := c.ValidateToken()
		if err != nil {
			return "", err
		}
		return u.Username, nil
	default:
		return "", fmt.Errorf("unknown provider")
	}
}
