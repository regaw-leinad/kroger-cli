package cmd

import (
	"fmt"
	"os"

	"github.com/regaw-leinad/kroger-cli/internal/config"
	"github.com/regaw-leinad/kroger-cli/internal/output"
	"github.com/spf13/cobra"
)

// Execute runs the root command and returns an exit code.
func Execute() int {
	root := newRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}
	return 0
}

func newRootCmd() *cobra.Command {
	var jsonFlag bool
	var quietFlag bool

	root := &cobra.Command{
		Use:   "kroger",
		Short: "Kroger CLI - search products, manage your cart, and shop with AI agents",
		Long: `A command-line interface for the Kroger API.

Search for products, find nearby stores, and add items to your cart.
Designed for both human use and AI agent automation.

Supports all Kroger-owned chains: Kroger, Fred Meyer, Ralphs, Harris Teeter,
King Soopers, Dillons, Fry's, QFC, Smith's, Jay C, Food 4 Less, FoodsCo,
Mariano's, Metro Market, Pick 'n Save, City Market, and more.

Get started:
  kroger auth setup     # Enter your Kroger API credentials
  kroger auth login     # Log in with your Kroger account
  kroger store search   # Find a nearby store
  kroger store select   # Set your default store`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var p output.Printer
			if jsonFlag || !output.IsTerminal() {
				p = output.NewJSONPrinter(os.Stdout, os.Stderr)
			} else {
				p = output.NewHumanPrinter(os.Stdout, os.Stderr)
			}
			if quietFlag {
				p = output.NewQuietPrinter(p)
			}
			cmd.SetContext(output.WithPrinter(cmd.Context(), p))
		},
	}

	root.CompletionOptions.HiddenDefaultCmd = true

	root.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Output as JSON (default when stdout is not a terminal)")
	root.PersistentFlags().BoolVar(&quietFlag, "quiet", false, "Suppress non-data output")

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load config: %v\n", err)
		cfg = config.Default()
	}

	root.AddCommand(newAuthCmd(cfg))
	root.AddCommand(newStoreCmd(cfg))
	root.AddCommand(newProductCmd(cfg))
	root.AddCommand(newCartCmd(cfg))

	return root
}
