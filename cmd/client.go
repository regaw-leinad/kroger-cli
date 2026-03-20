package cmd

import (
	"fmt"

	"github.com/regaw-leinad/kroger-cli/internal/api"
	"github.com/regaw-leinad/kroger-cli/internal/secrets"
)

// newAPIClient creates an authenticated API client.
func newAPIClient() (*api.Client, error) {
	store, err := secrets.Open()
	if err != nil {
		return nil, fmt.Errorf("opening secrets store: %w", err)
	}

	clientID, clientSecret, err := resolveCredentials(store)
	if err != nil {
		return nil, err
	}

	return api.NewClient(store, clientID, clientSecret), nil
}
