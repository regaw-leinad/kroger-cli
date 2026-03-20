package api

import "fmt"

// CartItem represents an item to add to the cart.
type CartItem struct {
	UPC      string `json:"upc"`
	Quantity int    `json:"quantity"`
}

type addToCartRequest struct {
	Items []CartItem `json:"items"`
}

// AddToCart adds items to the authenticated user's cart.
func (c *Client) AddToCart(items []CartItem) error {
	body := addToCartRequest{Items: items}
	if err := c.Put("/cart/add", body, nil); err != nil {
		return fmt.Errorf("adding to cart: %w", err)
	}
	return nil
}
