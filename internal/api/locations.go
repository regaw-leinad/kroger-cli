package api

import (
	"fmt"
	"net/url"
	"strconv"
)

// Location represents a Kroger store location.
type Location struct {
	LocationID  string          `json:"locationId"`
	Chain       string          `json:"chain"`
	Name        string          `json:"name"`
	Phone       string          `json:"phone"`
	Address     LocationAddress `json:"address"`
	Geolocation Geolocation     `json:"geolocation"`
	Hours       LocationHours   `json:"hours"`
	Departments []Department    `json:"departments"`
}

// LocationAddress holds a store's address.
type LocationAddress struct {
	AddressLine1 string `json:"addressLine1"`
	City         string `json:"city"`
	State        string `json:"state"`
	ZipCode      string `json:"zipCode"`
	County       string `json:"county"`
}

// FullAddress returns a single-line address string.
func (a LocationAddress) FullAddress() string {
	return fmt.Sprintf("%s, %s, %s %s", a.AddressLine1, a.City, a.State, a.ZipCode)
}

// Geolocation holds lat/lon coordinates.
type Geolocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// LocationHours holds store hours.
type LocationHours struct {
	Timezone  string `json:"timezone"`
	GMTOffset string `json:"gmtOffset"`
	Open24    bool   `json:"open24"`
}

// Department represents a store department.
type Department struct {
	DepartmentID string `json:"departmentId"`
	Name         string `json:"name"`
	Phone        string `json:"phone"`
}

type locationsResponse struct {
	Data []Location `json:"data"`
	Meta struct {
		Pagination Pagination `json:"pagination"`
	} `json:"meta"`
}

// Pagination holds API pagination info.
type Pagination struct {
	Start int `json:"start"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// SearchLocationsParams holds parameters for searching locations.
type SearchLocationsParams struct {
	ZipCode string
	Lat     float64
	Lon     float64
	Radius  int
	Chain   string
	Limit   int
}

// SearchLocations finds stores near a location.
func (c *Client) SearchLocations(p SearchLocationsParams) ([]Location, error) {
	params := url.Values{}

	if p.ZipCode != "" {
		params.Set("filter.zipCode.near", p.ZipCode)
	}
	if p.Lat != 0 {
		params.Set("filter.lat.near", strconv.FormatFloat(p.Lat, 'f', -1, 64))
	}
	if p.Lon != 0 {
		params.Set("filter.lon.near", strconv.FormatFloat(p.Lon, 'f', -1, 64))
	}
	if p.Radius > 0 {
		params.Set("filter.radiusInMiles", strconv.Itoa(p.Radius))
	}
	if p.Chain != "" {
		params.Set("filter.chain", p.Chain)
	}
	if p.Limit > 0 {
		params.Set("filter.limit", strconv.Itoa(p.Limit))
	}

	var resp locationsResponse
	if err := c.Get("/locations", params, &resp); err != nil {
		return nil, fmt.Errorf("searching locations: %w", err)
	}
	return resp.Data, nil
}

// GetLocation fetches a single location by ID.
func (c *Client) GetLocation(id string) (*Location, error) {
	var resp struct {
		Data Location `json:"data"`
	}
	if err := c.Get("/locations/"+id, nil, &resp); err != nil {
		return nil, fmt.Errorf("getting location: %w", err)
	}
	return &resp.Data, nil
}
