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

var tokenURLs = map[string]string{
	"vercel":  "https://vercel.com/account/tokens",
	"railway": "https://railway.com/account/tokens",
}

func authCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth <provider>",
		Short: "Authenticate with a cloud provider",
		Long: fmt.Sprintf("Save an API token for a cloud provider.\n\nSupported: %s",
			strings.Join(supportedProviders, ", ")),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := strings.ToLower(args[0])
			if !isSupportedProvider(provider) {
				return fmt.Errorf("unknown provider %q — supported: %s",
					provider, strings.Join(supportedProviders, ", "))
			}

			return runAuth(provider)
		},
	}

	return cmd
}

func runAuth(prov string) error {
	fmt.Printf("Create a token at: %s\n\n", tokenURLs[prov])

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
	case "railway":
		c := railway.NewClient(token)
		u, err := c.ValidateToken()
		if err != nil {
			return "", err
		}
		if u.Name != "" {
			return u.Name, nil
		}
		return u.Email, nil
	default:
		return "", fmt.Errorf("unknown provider")
	}
}
