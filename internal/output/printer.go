package output

import (
	"time"

	"github.com/regaw-leinad/kroger-cli/internal/api"
)

// AuthStatus holds the data needed to render auth status.
type AuthStatus struct {
	HasCredentials bool
	ClientID       string
	LoggedIn       bool
	TokenExpired   bool
	ExpiresAt      time.Time
	StoreID        string
	StoreName      string
}

// CartResult holds the result of adding items to cart.
type CartResult struct {
	Items []api.CartItem
}

// Printer defines how all CLI output is rendered.
// Commands return domain objects; the printer decides formatting.
type Printer interface {
	// Domain data rendering
	Products(products []api.Product, storeDomain string) error
	Product(product *api.Product, storeDomain string) error
	Locations(locations []api.Location) error
	Location(location *api.Location) error
	StoreSelected(id, name string) error
	AuthStatus(status AuthStatus) error
	CartAdded(result CartResult) error

	// Messages (stderr, suppressed in quiet mode)
	Status(format string, args ...any)
	Success(format string, args ...any)
}
