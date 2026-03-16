package commands

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/spf13/cobra"

	"github.com/lucasnevespereira/logmx/internal/auth"
	"github.com/lucasnevespereira/logmx/internal/provider/railway"
	"github.com/lucasnevespereira/logmx/internal/provider/vercel"
)

var supportedProviders = []string{"vercel", "railway"}

func authCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth <provider>",
		Short: "Authenticate with a cloud provider",
		Long: fmt.Sprintf("Authenticate with a cloud provider.\n\nSupported providers: %s",
			strings.Join(supportedProviders, ", ")),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := strings.ToLower(args[0])

			if !isSupportedProvider(provider) {
				return fmt.Errorf("unknown provider %q — supported: %s",
					provider, strings.Join(supportedProviders, ", "))
			}

			store := auth.NewStore()
			if err := store.Load(); err != nil {
				return err
			}

			fmt.Printf("Authenticate with %s\n", provider)
			fmt.Printf("Create a token at: %s\n\n", tokenURL(provider))
			fmt.Print("Paste your token: ")

			tokenBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				return fmt.Errorf("reading token: %w", err)
			}

			token := strings.TrimSpace(string(tokenBytes))
			if token == "" {
				return fmt.Errorf("token cannot be empty")
			}

			// Validate the token
			fmt.Print("Verifying... ")
			name, err := validateToken(provider, token)
			if err != nil {
				fmt.Println("failed")
				return fmt.Errorf("invalid token: %w", err)
			}
			fmt.Printf("authenticated as %s\n", name)

			store.Set(provider, token)
			if err := store.Save(); err != nil {
				return fmt.Errorf("saving token: %w", err)
			}

			fmt.Printf("\nRun 'logmx add' to pick projects from %s.\n", provider)
			return nil
		},
	}

	return cmd
}

func isSupportedProvider(p string) bool {
	for _, s := range supportedProviders {
		if s == p {
			return true
		}
	}
	return false
}

func tokenURL(provider string) string {
	switch provider {
	case "vercel":
		return "https://vercel.com/account/tokens"
	case "railway":
		return "https://railway.app/account/tokens"
	default:
		return ""
	}
}

func validateToken(provider, token string) (string, error) {
	switch provider {
	case "vercel":
		c := vercel.NewClient(token)
		u, err := c.GetUser()
		if err != nil {
			return "", err
		}
		return u.Username, nil
	case "railway":
		c := railway.NewClient(token)
		u, err := c.GetUser()
		if err != nil {
			return "", err
		}
		return u.Email, nil
	default:
		return "", fmt.Errorf("unknown provider")
	}
}
