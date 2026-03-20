package cmd

import (
	"fmt"

	"github.com/regaw-leinad/kroger-cli/internal/api"
	"github.com/regaw-leinad/kroger-cli/internal/config"
	"github.com/regaw-leinad/kroger-cli/internal/output"
	"github.com/spf13/cobra"
)

func newProductCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "product",
		Short: "Search and view products",
	}

	cmd.AddCommand(newProductSearchCmd(cfg))
	cmd.AddCommand(newProductShowCmd(cfg))

	return cmd
}

func newProductSearchCmd(cfg *config.Config) *cobra.Command {
	var brand string
	var limit int
	var fulfillment string

	cmd := &cobra.Command{
		Use:   "search <term>",
		Short: "Search for products",
		Example: `  kroger product search "organic milk"
  kroger product search "bread" --brand "Dave's Killer Bread" --limit 5
  kroger product search "chicken" --fulfillment pickup`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.From(cmd.Context())

			if cfg.DefaultStoreID == "" {
				return fmt.Errorf("no default store set - run 'kroger store select <id>' first (needed for pricing)")
			}

			client, err := newAPIClient()
			if err != nil {
				return err
			}

			products, err := client.SearchProducts(api.SearchProductsParams{
				Term:        args[0],
				Brand:       brand,
				LocationID:  cfg.DefaultStoreID,
				Fulfillment: fulfillment,
				Limit:       limit,
			})
			if err != nil {
				return err
			}

			if len(products) == 0 {
				out.Status("No products found for %q", args[0])
				return nil
			}

			return out.Products(products, cfg.StoreDomain())
		},
	}

	cmd.Flags().StringVar(&brand, "brand", "", "Filter by brand name")
	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum results (1-50)")
	cmd.Flags().StringVar(&fulfillment, "fulfillment", "", "Filter by fulfillment type (pickup, delivery, in_store, ship)")

	return cmd
}

func newProductShowCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "show <product-id>",
		Short: "Show product details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.From(cmd.Context())

			client, err := newAPIClient()
			if err != nil {
				return err
			}

			product, err := client.GetProduct(args[0], cfg.DefaultStoreID)
			if err != nil {
				return err
			}

			return out.Product(product, cfg.StoreDomain())
		},
	}
}
