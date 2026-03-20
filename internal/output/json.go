package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/regaw-leinad/kroger-cli/internal/api"
)

// JSONPrinter renders all output as JSON.
type JSONPrinter struct {
	out io.Writer
	err io.Writer
}

func NewJSONPrinter(out, err io.Writer) *JSONPrinter {
	return &JSONPrinter{out: out, err: err}
}

func (p *JSONPrinter) Products(products []api.Product, storeDomain string) error {
	slim := make([]api.SlimProduct, len(products))
	for i, prod := range products {
		slim[i] = prod.Slim(storeDomain)
	}
	return p.encode(slim)
}

func (p *JSONPrinter) Product(product *api.Product, storeDomain string) error {
	return p.encode(product.Slim(storeDomain))
}

func (p *JSONPrinter) Locations(locations []api.Location) error {
	return p.encode(locations)
}

func (p *JSONPrinter) Location(location *api.Location) error {
	return p.encode(location)
}

func (p *JSONPrinter) StoreSelected(id, name string) error {
	return p.encode(map[string]string{
		"store_id":   id,
		"store_name": name,
	})
}

func (p *JSONPrinter) AuthStatus(status AuthStatus) error {
	return p.encode(map[string]any{
		"has_credentials": status.HasCredentials,
		"logged_in":       status.LoggedIn,
		"token_expired":   status.TokenExpired,
		"expires_at":      status.ExpiresAt,
		"store_id":        status.StoreID,
		"store_name":      status.StoreName,
	})
}

func (p *JSONPrinter) CartAdded(result CartResult) error {
	return p.encode(map[string]any{
		"added": result.Items,
		"count": len(result.Items),
	})
}

func (p *JSONPrinter) Status(format string, args ...any) {
	fmt.Fprintf(p.err, format+"\n", args...)
}

func (p *JSONPrinter) Success(format string, args ...any) {
	p.Status("✓ "+format, args...)
}

func (p *JSONPrinter) encode(v any) error {
	enc := json.NewEncoder(p.out)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
