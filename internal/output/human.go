package output

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/regaw-leinad/kroger-cli/internal/api"
)

var (
	green   = lipgloss.Color("42")
	yellow  = lipgloss.Color("214")
	red     = lipgloss.Color("196")
	purple  = lipgloss.Color("99")
	dimGray = lipgloss.Color("248")
	border  = lipgloss.Color("63")

	priceStyle = lipgloss.NewStyle().Foreground(green).Bold(true)
	promoStyle = lipgloss.NewStyle().Foreground(yellow).Bold(true)
	oldPrice   = lipgloss.NewStyle().Foreground(dimGray).Strikethrough(true)
	dimStyle   = lipgloss.NewStyle().Foreground(dimGray)
	boldStyle  = lipgloss.NewStyle().Bold(true)
	errStyle   = lipgloss.NewStyle().Foreground(red).Bold(true)
	okStyle    = lipgloss.NewStyle().Foreground(green).Bold(true)
	labelStyle = lipgloss.NewStyle().Foreground(dimGray)
)

// HumanPrinter renders output as styled terminal text and tables.
type HumanPrinter struct {
	out io.Writer
	err io.Writer
}

func NewHumanPrinter(out, err io.Writer) *HumanPrinter {
	return &HumanPrinter{out: out, err: err}
}

func (p *HumanPrinter) Products(products []api.Product, storeDomain string) error {
	rows := make([][]string, 0, len(products))
	for _, prod := range products {
		size := ""
		price := ""
		if len(prod.Items) > 0 {
			item := prod.Items[0]
			size = item.Size
			if item.Price.Promo > 0 {
				price = promoStyle.Render(fmt.Sprintf("$%.2f", item.Price.Promo)) +
					" " + oldPrice.Render(fmt.Sprintf("$%.2f", item.Price.Regular))
			} else if item.Price.Regular > 0 {
				price = priceStyle.Render(fmt.Sprintf("$%.2f", item.Price.Regular))
			}
		}
		rows = append(rows, []string{
			prod.UPC,
			dimStyle.Render(prod.Brand),
			boldStyle.Render(prod.Description),
			size,
			price,
		})
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(border)).
		StyleFunc(func(row, col int) lipgloss.Style {
			s := lipgloss.NewStyle().Padding(0, 1)
			if row == table.HeaderRow {
				return s.Bold(true).Foreground(purple)
			}
			return s
		}).
		Headers("UPC", "BRAND", "DESCRIPTION", "SIZE", "PRICE").
		Rows(rows...)

	fmt.Fprintln(p.out, t)
	return nil
}

func (p *HumanPrinter) Product(product *api.Product, storeDomain string) error {
	fmt.Fprintf(p.out, "%s %s\n", labelStyle.Render("Product:"), boldStyle.Render(product.Description))
	fmt.Fprintf(p.out, "%s %s\n", labelStyle.Render("Brand:"), product.Brand)
	fmt.Fprintf(p.out, "%s %s\n", labelStyle.Render("UPC:"), product.UPC)
	fmt.Fprintf(p.out, "%s %s\n", labelStyle.Render("URL:"), product.ProductURL(storeDomain))

	if len(product.Categories) > 0 {
		fmt.Fprintf(p.out, "%s %s\n", labelStyle.Render("Category:"), product.Categories[0])
	}

	for _, item := range product.Items {
		fmt.Fprintf(p.out, "%s %s\n", labelStyle.Render("Size:"), item.Size)
		if item.Price.Promo > 0 {
			fmt.Fprintf(p.out, "%s %s %s\n",
				labelStyle.Render("Price:"),
				promoStyle.Render(fmt.Sprintf("$%.2f", item.Price.Promo)),
				oldPrice.Render(fmt.Sprintf("$%.2f", item.Price.Regular)))
		} else if item.Price.Regular > 0 {
			fmt.Fprintf(p.out, "%s %s\n",
				labelStyle.Render("Price:"),
				priceStyle.Render(fmt.Sprintf("$%.2f", item.Price.Regular)))
		}
		fmt.Fprintf(p.out, "%s %s  %s %s  %s %s\n",
			labelStyle.Render("Pickup:"), fmtBool(item.Fulfillment.Curbside),
			labelStyle.Render("Delivery:"), fmtBool(item.Fulfillment.Delivery),
			labelStyle.Render("In-Store:"), fmtBool(item.Fulfillment.InStore))
	}
	return nil
}

func (p *HumanPrinter) Locations(locations []api.Location) error {
	rows := make([][]string, 0, len(locations))
	for _, loc := range locations {
		rows = append(rows, []string{
			loc.LocationID,
			boldStyle.Render(loc.Name),
			loc.Address.FullAddress(),
			loc.Phone,
		})
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(border)).
		StyleFunc(func(row, col int) lipgloss.Style {
			s := lipgloss.NewStyle().Padding(0, 1)
			if row == table.HeaderRow {
				return s.Bold(true).Foreground(purple)
			}
			return s
		}).
		Headers("ID", "NAME", "ADDRESS", "PHONE").
		Rows(rows...)

	fmt.Fprintln(p.out, t)
	return nil
}

func (p *HumanPrinter) Location(location *api.Location, showDepartments bool) error {
	fmt.Fprintf(p.out, "%s %s\n", labelStyle.Render("Name:"), boldStyle.Render(location.Name))
	fmt.Fprintf(p.out, "%s %s\n", labelStyle.Render("ID:"), location.LocationID)
	fmt.Fprintf(p.out, "%s %s\n", labelStyle.Render("Chain:"), location.Chain)
	fmt.Fprintf(p.out, "%s %s\n", labelStyle.Render("Address:"), location.Address.FullAddress())
	fmt.Fprintf(p.out, "%s %s\n", labelStyle.Render("Phone:"), location.Phone)

	slim := location.Slim()
	fmt.Fprintf(p.out, "%s %s  %s %s\n",
		labelStyle.Render("Pickup:"), fmtBool(slim.HasPickup),
		labelStyle.Render("Pharmacy:"), fmtBool(slim.HasPharmacy))

	if showDepartments && len(location.Departments) > 0 {
		fmt.Fprintln(p.out)
		fmt.Fprintln(p.out, boldStyle.Render("Departments:"))
		for _, dept := range location.Departments {
			fmt.Fprintf(p.out, "  %s %s\n", dept.Name, dimStyle.Render("("+dept.DepartmentID+")"))
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
		fmt.Fprintf(p.out, "%s %s\n", labelStyle.Render("API credentials:"), errStyle.Render("not configured"))
		fmt.Fprintf(p.out, "  Run: %s\n", boldStyle.Render("kroger auth setup"))
		return nil
	}
	cid := status.ClientID
	if len(cid) > 12 {
		cid = cid[:12]
	}
	fmt.Fprintf(p.out, "%s %s\n", labelStyle.Render("API credentials:"), okStyle.Render("configured")+" "+dimStyle.Render("("+cid+"...)"))

	if !status.LoggedIn {
		fmt.Fprintf(p.out, "%s %s\n", labelStyle.Render("Logged in:"), errStyle.Render("no"))
		fmt.Fprintf(p.out, "  Run: %s\n", boldStyle.Render("kroger auth login"))
		return nil
	}

	if status.TokenExpired {
		fmt.Fprintf(p.out, "%s %s\n", labelStyle.Render("Logged in:"),
			lipgloss.NewStyle().Foreground(yellow).Render("yes (token expired, will refresh)"))
	} else {
		fmt.Fprintf(p.out, "%s %s %s\n", labelStyle.Render("Logged in:"),
			okStyle.Render("yes"),
			dimStyle.Render("(expires "+status.ExpiresAt.Local().Format("3:04 PM")+")"))
	}

	if status.StoreID != "" {
		fmt.Fprintf(p.out, "%s %s %s\n", labelStyle.Render("Default store:"),
			boldStyle.Render(status.StoreName),
			dimStyle.Render("("+status.StoreID+")"))
	} else {
		fmt.Fprintf(p.out, "%s %s\n", labelStyle.Render("Default store:"), errStyle.Render("not set"))
		fmt.Fprintf(p.out, "  Run: %s\n", boldStyle.Render("kroger store search --zip <zip>"))
	}
	return nil
}

func (p *HumanPrinter) CartAdded(result CartResult) error {
	for _, item := range result.Items {
		p.Success("Added %s (qty: %d) to cart", item.UPC, item.Quantity)
	}
	fmt.Fprintf(p.err, "%s\n", dimStyle.Render("Complete your order at kroger.com"))
	return nil
}

func (p *HumanPrinter) Status(format string, args ...any) {
	fmt.Fprintf(p.err, format+"\n", args...)
}

func (p *HumanPrinter) Success(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(p.err, "%s %s\n", okStyle.Render("✓"), msg)
}

func fmtBool(b bool) string {
	if b {
		return okStyle.Render("yes")
	}
	return errStyle.Render("no")
}
