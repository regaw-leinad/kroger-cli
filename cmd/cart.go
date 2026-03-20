package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/regaw-leinad/kroger-cli/internal/api"
	"github.com/regaw-leinad/kroger-cli/internal/config"
	"github.com/regaw-leinad/kroger-cli/internal/output"
	"github.com/spf13/cobra"
)

func newCartCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cart",
		Short: "Manage your Kroger cart",
	}

	cmd.AddCommand(newCartAddCmd())

	return cmd
}

func newCartAddCmd() *cobra.Command {
	var quantity int
	var itemsJSON string

	cmd := &cobra.Command{
		Use:   "add [upc]",
		Short: "Add items to your cart",
		Long: `Add one or more items to your Kroger cart by UPC.

For a single item, pass the UPC as an argument:
  kroger cart add 0001111041600 --quantity 2

For multiple items, use --items with JSON:
  kroger cart add --items '[{"upc":"0001111041600","quantity":2},{"upc":"0001111041700","quantity":1}]'

Note: The Kroger API only supports adding items. You cannot view or remove cart items via the API.
Complete your order at kroger.com.`,
		Example: `  kroger cart add 0001111041600
  kroger cart add 0001111041600 --quantity 3
  kroger cart add --items '[{"upc":"0001111041600","quantity":2}]'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.From(cmd.Context())

			var items []api.CartItem

			if itemsJSON != "" {
				if err := json.Unmarshal([]byte(itemsJSON), &items); err != nil {
					return fmt.Errorf("parsing --items JSON: %w", err)
				}
			} else if len(args) > 0 {
				items = []api.CartItem{{UPC: args[0], Quantity: quantity}}
			} else {
				return fmt.Errorf("provide a UPC argument or --items JSON")
			}

			for i := range items {
				if items[i].Quantity < 1 {
					items[i].Quantity = 1
				}
			}

			client, err := newAPIClient()
			if err != nil {
				return err
			}

			if err := client.AddToCart(items); err != nil {
				return err
			}

			return out.CartAdded(output.CartResult{Items: items})
		},
	}

	cmd.Flags().IntVar(&quantity, "quantity", 1, "Quantity to add (single item mode)")
	cmd.Flags().StringVar(&itemsJSON, "items", "", "JSON array of items to add")

	return cmd
}
