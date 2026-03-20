package output

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/regaw-leinad/kroger-cli/internal/api"
)

// HumanPrinter renders output as human-readable text and tables.
type HumanPrinter struct {
	out io.Writer
	err io.Writer
}

func NewHumanPrinter(out, err io.Writer) *HumanPrinter {
	return &HumanPrinter{out: out, err: err}
}

func (p *HumanPrinter) Products(products []api.Product, storeDomain string) error {
	tw := tabwriter.NewWriter(p.out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "UPC\tBRAND\tDESCRIPTION\tSIZE\tPRICE")
	for _, prod := range products {
		size := ""
		price := ""
		if len(prod.Items) > 0 {
			item := prod.Items[0]
			size = item.Size
			if item.Price.Promo > 0 {
				price = fmt.Sprintf("$%.2f (was $%.2f)", item.Price.Promo, item.Price.Regular)
			} else if item.Price.Regular > 0 {
				price = fmt.Sprintf("$%.2f", item.Price.Regular)
			}
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			prod.UPC, prod.Brand, prod.Description, size, price,
		)
	}
	return tw.Flush()
}

func (p *HumanPrinter) Product(product *api.Product, storeDomain string) error {
	fmt.Fprintf(p.out, "Product: %s\n", product.Description)
	fmt.Fprintf(p.out, "Brand:   %s\n", product.Brand)
	fmt.Fprintf(p.out, "UPC:     %s\n", product.UPC)
	fmt.Fprintf(p.out, "URL:     %s\n", product.ProductURL(storeDomain))

	if len(product.Categories) > 0 {
		fmt.Fprintf(p.out, "Category: %s\n", product.Categories[0])
	}

	for _, item := range product.Items {
		fmt.Fprintf(p.out, "\nSize:    %s\n", item.Size)
		if item.Price.Promo > 0 {
			fmt.Fprintf(p.out, "Price:   $%.2f (regular $%.2f)\n", item.Price.Promo, item.Price.Regular)
		} else if item.Price.Regular > 0 {
			fmt.Fprintf(p.out, "Price:   $%.2f\n", item.Price.Regular)
		}
		fmt.Fprintf(p.out, "Pickup:  %v  Delivery: %v  In-Store: %v\n",
			item.Fulfillment.Curbside, item.Fulfillment.Delivery, item.Fulfillment.InStore)
	}
	return nil
}

func (p *HumanPrinter) Locations(locations []api.Location) error {
	tw := tabwriter.NewWriter(p.out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tNAME\tADDRESS\tPHONE")
	for _, loc := range locations {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			loc.LocationID, loc.Name, loc.Address.FullAddress(), loc.Phone,
		)
	}
	return tw.Flush()
}

func (p *HumanPrinter) Location(location *api.Location) error {
	fmt.Fprintf(p.out, "Name:    %s\n", location.Name)
	fmt.Fprintf(p.out, "ID:      %s\n", location.LocationID)
	fmt.Fprintf(p.out, "Chain:   %s\n", location.Chain)
	fmt.Fprintf(p.out, "Address: %s\n", location.Address.FullAddress())
	fmt.Fprintf(p.out, "Phone:   %s\n", location.Phone)

	if len(location.Departments) > 0 {
		fmt.Fprintln(p.out, "\nDepartments:")
		for _, dept := range location.Departments {
			fmt.Fprintf(p.out, "  %s (%s)\n", dept.Name, dept.DepartmentID)
		}
	}
	return nil
}

func (p *HumanPrinter) StoreSelected(id, name string) error {
	p.Success("Default store: %s (%s)", name, id)
	return nil
}

func (p *HumanPrinter) AuthStatus(status AuthStatus) error {
	if !status.HasCredentials {
		fmt.Fprintln(p.out, "API credentials: not configured")
		fmt.Fprintln(p.out, "  Run: kroger auth setup")
		return nil
	}
	cid := status.ClientID
	if len(cid) > 12 {
		cid = cid[:12]
	}
	fmt.Fprintf(p.out, "API credentials: configured (client_id: %s...)\n", cid)

	if !status.LoggedIn {
		fmt.Fprintln(p.out, "Logged in: no")
		fmt.Fprintln(p.out, "  Run: kroger auth login")
		return nil
	}

	if status.TokenExpired {
		fmt.Fprintln(p.out, "Logged in: yes (token expired, will refresh on next request)")
	} else {
		fmt.Fprintf(p.out, "Logged in: yes (token expires at %s)\n", status.ExpiresAt)
	}

	if status.StoreID != "" {
		fmt.Fprintf(p.out, "Default store: %s (%s)\n", status.StoreName, status.StoreID)
	} else {
		fmt.Fprintln(p.out, "Default store: not set")
		fmt.Fprintln(p.out, "  Run: kroger store search --zip <zip> && kroger store select <id>")
	}
	return nil
}

func (p *HumanPrinter) CartAdded(result CartResult) error {
	for _, item := range result.Items {
		p.Success("Added %s (qty: %d) to cart", item.UPC, item.Quantity)
	}
	p.Status("Complete your order at kroger.com")
	return nil
}

func (p *HumanPrinter) Status(format string, args ...any) {
	fmt.Fprintf(p.err, format+"\n", args...)
}

func (p *HumanPrinter) Success(format string, args ...any) {
	p.Status("✓ "+format, args...)
}
