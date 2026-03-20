package cmd

import (
	"fmt"

	"github.com/regaw-leinad/kroger-cli/internal/api"
	"github.com/regaw-leinad/kroger-cli/internal/config"
	"github.com/regaw-leinad/kroger-cli/internal/output"
	"github.com/spf13/cobra"
)

func newStoreCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store",
		Short: "Find and manage Kroger stores",
	}

	cmd.AddCommand(newStoreSearchCmd(cfg))
	cmd.AddCommand(newStoreSelectCmd(cfg))
	cmd.AddCommand(newStoreShowCmd(cfg))

	return cmd
}

func newStoreSearchCmd(cfg *config.Config) *cobra.Command {
	var zip string
	var lat, lon float64
	var radius, limit int
	var chain string

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for nearby stores",
		Example: `  kroger store search --zip 45202
  kroger store search --zip 45202 --radius 20 --chain Kroger
  kroger store search --lat 39.1 --lon -84.5`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.From(cmd.Context())

			if zip == "" && lat == 0 && lon == 0 {
				return fmt.Errorf("provide --zip or --lat/--lon to search")
			}

			client, err := newAPIClient()
			if err != nil {
				return err
			}

			locations, err := client.SearchLocations(api.SearchLocationsParams{
				ZipCode: zip,
				Lat:     lat,
				Lon:     lon,
				Radius:  radius,
				Chain:   chain,
				Limit:   limit,
			})
			if err != nil {
				return err
			}

			if len(locations) == 0 {
				out.Status("No stores found")
				return nil
			}

			return out.Locations(locations)
		},
	}

	cmd.Flags().StringVar(&zip, "zip", "", "Zip code to search near")
	cmd.Flags().Float64Var(&lat, "lat", 0, "Latitude")
	cmd.Flags().Float64Var(&lon, "lon", 0, "Longitude")
	cmd.Flags().IntVar(&radius, "radius", 10, "Search radius in miles")
	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum results")
	cmd.Flags().StringVar(&chain, "chain", "", "Filter by chain (Kroger, Ralphs, \"Harris Teeter\", \"Fred Meyer\", \"King Soopers\", etc.)")

	return cmd
}

func newStoreSelectCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "select <location-id>",
		Short: "Set default store for product searches",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.From(cmd.Context())

			client, err := newAPIClient()
			if err != nil {
				return err
			}

			loc, err := client.GetLocation(args[0])
			if err != nil {
				return fmt.Errorf("fetching store details: %w", err)
			}

			if err := cfg.SetStore(loc.LocationID, loc.Name, loc.Chain); err != nil {
				return fmt.Errorf("saving store selection: %w", err)
			}

			return out.StoreSelected(loc.LocationID, loc.Name)
		},
	}
}

func newStoreShowCmd(cfg *config.Config) *cobra.Command {
	var showDepartments bool

	cmd := &cobra.Command{
		Use:   "show [location-id]",
		Short: "Show store details",
		Long:  "Show details for a store. If no ID is given, shows the default store.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.From(cmd.Context())

			locationID := cfg.DefaultStoreID
			if len(args) > 0 {
				locationID = args[0]
			}
			if locationID == "" {
				return fmt.Errorf("no store specified and no default store set - run 'kroger store select <id>'")
			}

			client, err := newAPIClient()
			if err != nil {
				return err
			}

			loc, err := client.GetLocation(locationID)
			if err != nil {
				return err
			}

			return out.Location(loc, showDepartments)
		},
	}

	cmd.Flags().BoolVar(&showDepartments, "departments", false, "Show full department list")

	return cmd
}
