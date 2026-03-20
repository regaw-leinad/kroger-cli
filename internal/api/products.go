package api

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// Product represents a Kroger product.
type Product struct {
	ProductID   string         `json:"productId"`
	UPC         string         `json:"upc"`
	Brand       string         `json:"brand"`
	Categories  []string       `json:"categories"`
	Description string         `json:"description"`
	Images      []ProductImage `json:"images"`
	Items       []ProductItem  `json:"items"`
	Temperature *Temperature   `json:"temperature,omitempty"`
}

// ProductImage holds image data.
type ProductImage struct {
	Perspective string      `json:"perspective"`
	Featured    bool        `json:"featured"`
	Sizes       []ImageSize `json:"sizes"`
}

// ImageSize holds a single image URL and size.
type ImageSize struct {
	Size string `json:"size"`
	URL  string `json:"url"`
}

// ProductItem holds item-level details (pricing, fulfillment).
type ProductItem struct {
	ItemID      string      `json:"itemId"`
	Favorite    bool        `json:"favorite"`
	Fulfillment Fulfillment `json:"fulfillment"`
	Price       Price       `json:"price"`
	Size        string      `json:"size"`
	SoldBy      string      `json:"soldBy"`
}

// Fulfillment holds availability info.
type Fulfillment struct {
	Curbside   bool `json:"curbside"`
	Delivery   bool `json:"delivery"`
	InStore    bool `json:"inStore"`
	ShipToHome bool `json:"shipToHome"`
}

// Price holds pricing info.
type Price struct {
	Regular float64 `json:"regular"`
	Promo   float64 `json:"promo"`
}

// Temperature holds temperature sensitivity info.
type Temperature struct {
	Indicator     string `json:"indicator"`
	HeatSensitive bool   `json:"heatSensitive"`
}

// DisplayPrice returns the best price for display.
func (p Price) DisplayPrice() float64 {
	if p.Promo > 0 {
		return p.Promo
	}
	return p.Regular
}

// SlimFulfillment holds availability without shipToHome (always false).
type SlimFulfillment struct {
	Curbside bool `json:"curbside"`
	Delivery bool `json:"delivery"`
	InStore  bool `json:"inStore"`
}

// SlimProduct is a flattened, agent-friendly product representation.
type SlimProduct struct {
	UPC         string          `json:"upc"`
	Brand       string          `json:"brand,omitempty"`
	Description string          `json:"description"`
	Size        string          `json:"size,omitempty"`
	Price       float64         `json:"price"`
	Promo       float64         `json:"promo,omitempty"`
	SoldBy      string          `json:"soldBy,omitempty"`
	Categories  []string        `json:"categories,omitempty"`
	Fulfillment SlimFulfillment `json:"fulfillment"`
	Temperature string          `json:"temperature,omitempty"`
	Image       string          `json:"image,omitempty"`
	URL         string          `json:"url"`
}

// Slim converts a Product to a SlimProduct.
func (p Product) Slim(storeDomain string) SlimProduct {
	s := SlimProduct{
		UPC:         p.UPC,
		Brand:       p.Brand,
		Description: p.Description,
		Categories:  dedup(p.Categories),
		Image:       p.FrontImageURL(),
		URL:         p.ProductURL(storeDomain),
	}
	if p.Temperature != nil {
		s.Temperature = p.Temperature.Indicator
	}
	if len(p.Items) > 0 {
		item := p.Items[0]
		s.Size = item.Size
		s.Price = item.Price.Regular
		if item.Price.Promo > 0 {
			s.Promo = item.Price.Promo
		}
		s.SoldBy = item.SoldBy
		s.Fulfillment = SlimFulfillment{
			Curbside: item.Fulfillment.Curbside,
			Delivery: item.Fulfillment.Delivery,
			InStore:  item.Fulfillment.InStore,
		}
	}
	return s
}

func dedup(s []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(s))
	for _, v := range s {
		if seen[v] {
			continue
		}
		seen[v] = true
		result = append(result, v)
	}
	return result
}

// FrontImageURL returns the medium front image URL, or empty string if unavailable.
func (p Product) FrontImageURL() string {
	for _, img := range p.Images {
		if img.Perspective != "front" {
			continue
		}
		for _, size := range img.Sizes {
			if size.Size == "xlarge" {
				return size.URL
			}
		}
		// fallback to first available size
		if len(img.Sizes) > 0 {
			return img.Sizes[0].URL
		}
	}
	// fallback to any featured image
	for _, img := range p.Images {
		if img.Featured && len(img.Sizes) > 0 {
			for _, size := range img.Sizes {
				if size.Size == "xlarge" {
					return size.URL
				}
			}
			return img.Sizes[0].URL
		}
	}
	return ""
}

var slugRe = regexp.MustCompile(`[^a-z0-9-]+`)
var multiDash = regexp.MustCompile(`-{2,}`)

// Slugify converts a product description to a URL-friendly slug.
func Slugify(s string) string {
	s = strings.ToLower(s)
	s = slugRe.ReplaceAllString(s, "-")
	s = multiDash.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// ProductURL returns the direct product page URL for a given store domain.
func (p Product) ProductURL(storeDomain string) string {
	slug := Slugify(p.Description)
	return fmt.Sprintf("https://%s/p/%s/%s", storeDomain, slug, p.UPC)
}

type productsResponse struct {
	Data []Product `json:"data"`
	Meta struct {
		Pagination Pagination `json:"pagination"`
	} `json:"meta"`
}

// SearchProductsParams holds parameters for searching products.
type SearchProductsParams struct {
	Term        string
	Brand       string
	LocationID  string
	Fulfillment string
	Limit       int
	Start       int
}

// SearchProducts searches for products.
func (c *Client) SearchProducts(p SearchProductsParams) ([]Product, error) {
	params := url.Values{}

	if p.Term != "" {
		params.Set("filter.term", p.Term)
	}
	if p.Brand != "" {
		params.Set("filter.brand", p.Brand)
	}
	if p.LocationID != "" {
		params.Set("filter.locationId", p.LocationID)
	}
	if p.Fulfillment != "" {
		params.Set("filter.fulfillment", p.Fulfillment)
	}
	if p.Limit > 0 {
		params.Set("filter.limit", strconv.Itoa(p.Limit))
	}
	if p.Start > 0 {
		params.Set("filter.start", strconv.Itoa(p.Start))
	}

	var resp productsResponse
	if err := c.Get("/products", params, &resp); err != nil {
		return nil, fmt.Errorf("searching products: %w", err)
	}
	return resp.Data, nil
}

// GetProduct fetches a single product by ID.
func (c *Client) GetProduct(id, locationID string) (*Product, error) {
	params := url.Values{}
	if locationID != "" {
		params.Set("filter.locationId", locationID)
	}

	var resp struct {
		Data Product `json:"data"`
	}
	if err := c.Get("/products/"+id, params, &resp); err != nil {
		return nil, fmt.Errorf("getting product: %w", err)
	}
	return &resp.Data, nil
}
