package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/regaw-leinad/kroger-cli/internal/auth"
	"github.com/regaw-leinad/kroger-cli/internal/config"
	"github.com/regaw-leinad/kroger-cli/internal/output"
	"github.com/regaw-leinad/kroger-cli/internal/secrets"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newAuthCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
	}

	cmd.AddCommand(newAuthSetupCmd(cfg))
	cmd.AddCommand(newAuthLoginCmd(cfg))
	cmd.AddCommand(newAuthStatusCmd(cfg))
	cmd.AddCommand(newAuthLogoutCmd())

	return cmd
}

func newAuthSetupCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Configure Kroger API credentials",
		Long: `Set up your Kroger API credentials for the first time.

You need a free Kroger Developer account. Follow the setup guide:
  https://github.com/regaw-leinad/kroger-cli/blob/main/docs/setup.md`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.From(cmd.Context())

			store, err := secrets.Open()
			if err != nil {
				return fmt.Errorf("opening secrets store: %w", err)
			}

			out.Status("First, create a Kroger Developer app by following the guide:")
			out.Status("  https://github.com/regaw-leinad/kroger-cli/blob/main/docs/setup.md")
			out.Status("")
			out.Status("Make sure to set the redirect URI to: %s", auth.RedirectURI)
			out.Status("")

			reader := bufio.NewReader(os.Stdin)

			fmt.Fprint(os.Stderr, "Client ID: ")
			clientID, _ := reader.ReadString('\n')
			clientID = strings.TrimSpace(clientID)
			if clientID == "" {
				return fmt.Errorf("client ID is required")
			}

			fmt.Fprint(os.Stderr, "Client Secret: ")
			secretBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Fprintln(os.Stderr)
			if err != nil {
				return fmt.Errorf("reading secret: %w", err)
			}
			clientSecret := strings.TrimSpace(string(secretBytes))
			if clientSecret == "" {
				return fmt.Errorf("client secret is required")
			}

			creds := secrets.Credentials{
				ClientID:     clientID,
				ClientSecret: clientSecret,
			}
			if err := store.SetCredentials(creds); err != nil {
				return fmt.Errorf("saving credentials: %w", err)
			}

			out.Success("Credentials saved")
			out.Status("")
			out.Status("Next, run: kroger auth login")
			return nil
		},
	}
}

func newAuthLoginCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Log in with your Kroger account",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.From(cmd.Context())

			store, err := secrets.Open()
			if err != nil {
				return fmt.Errorf("opening secrets store: %w", err)
			}

			clientID, clientSecret, err := resolveCredentials(store)
			if err != nil {
				return err
			}

			out.Status("Opening browser for Kroger login...")

			if err := auth.Login(cmd.Context(), store, clientID, clientSecret); err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			out.Success("Authenticated successfully")

			if cfg.DefaultStoreID == "" {
				out.Status("")
				out.Status("Next, find and select your preferred store:")
				out.Status("  kroger store search --zip <your-zip>")
				out.Status("  kroger store select <store-id>")
			}
			return nil
		},
	}
}

func newAuthStatusCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.From(cmd.Context())

			store, err := secrets.Open()
			if err != nil {
				return fmt.Errorf("opening secrets store: %w", err)
			}

			creds, err := store.GetCredentials()
			if err != nil {
				return fmt.Errorf("reading credentials: %w", err)
			}

			tok, err := store.GetToken()
			if err != nil {
				return fmt.Errorf("reading token: %w", err)
			}

			return out.AuthStatus(output.AuthStatus{
				HasCredentials: creds.ClientID != "",
				ClientID:       creds.ClientID,
				LoggedIn:       tok.AccessToken != "",
				TokenExpired:   tok.AccessToken != "" && tok.IsExpired(),
				ExpiresAt:      tok.ExpiresAt.Local().Format("3:04 PM"),
				StoreID:        cfg.DefaultStoreID,
				StoreName:      cfg.DefaultStoreName,
			})
		},
	}
}

func newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear stored credentials and tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.From(cmd.Context())

			store, err := secrets.Open()
			if err != nil {
				return fmt.Errorf("opening secrets store: %w", err)
			}

			if err := store.DeleteToken(); err != nil {
				return fmt.Errorf("deleting token: %w", err)
			}
			if err := store.DeleteCredentials(); err != nil {
				return fmt.Errorf("deleting credentials: %w", err)
			}

			out.Success("Logged out and credentials cleared")
			return nil
		},
	}
}

// resolveCredentials gets client credentials from env vars or the secrets store.
func resolveCredentials(store *secrets.Store) (clientID, clientSecret string, err error) {
	clientID = os.Getenv("KROGER_CLIENT_ID")
	clientSecret = os.Getenv("KROGER_CLIENT_SECRET")

	if clientID != "" && clientSecret != "" {
		return clientID, clientSecret, nil
	}

	creds, err := store.GetCredentials()
	if err != nil {
		return "", "", fmt.Errorf("reading credentials: %w", err)
	}
	if creds.ClientID == "" {
		return "", "", fmt.Errorf("no credentials configured - run 'kroger auth setup' first")
	}

	if clientID == "" {
		clientID = creds.ClientID
	}
	if clientSecret == "" {
		clientSecret = creds.ClientSecret
	}
	return clientID, clientSecret, nil
}
