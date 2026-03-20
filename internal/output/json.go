package output

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/regaw-leinad/kroger-cli/internal/api"
)

// JSONPrinter renders all output as JSON.
type JSONPrinter struct {
	enc *json.Encoder
	err io.Writer
}

func NewJSONPrinter(out, err io.Writer) *JSONPrinter {
	enc := json.NewEncoder(out)
	enc.SetEscapeHTML(false)
	return &JSONPrinter{enc: enc, err: err}
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
	slim := make([]api.SlimLocation, len(locations))
	for i, loc := range locations {
		slim[i] = loc.Slim()
	}
	return p.encode(slim)
}

func (p *JSONPrinter) Location(location *api.Location) error {
	return p.encode(location.Slim())
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
		"expires_at":      status.ExpiresAt.Format(time.RFC3339),
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
	return p.enc.Encode(v)
}
