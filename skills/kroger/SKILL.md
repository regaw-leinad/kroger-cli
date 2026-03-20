---
name: kroger
description: Shop for groceries at Kroger. Use when the user asks to search for groceries, add items to their Kroger cart, find a Kroger store, or do grocery shopping.
allowed-tools: Bash(kroger *)
---

# Kroger Shopping Assistant

Help users shop for groceries using the `kroger` CLI.

## Prerequisites

Before shopping, verify the CLI is set up:

```bash
kroger auth status --json
```

If not configured, guide the user through:
1. `kroger auth setup` - enter API credentials
2. `kroger auth login` - authenticate with Kroger
3. `kroger store search --zip <zip>` - find nearby stores
4. `kroger store select <id>` - set default store

## Shopping Workflow

### 1. Search for products

```bash
kroger product search "organic whole milk" --limit 10 --json
```

Flags: `--brand`, `--limit` (1-50), `--fulfillment` (pickup, delivery, in_store, ship).

### 2. Pick the best match

Review the results. Consider: price, brand preference, size, and availability. The UPC field identifies each product.

### 3. Add to cart

Single item:
```bash
kroger cart add <upc> --quantity 1 --json
```

Multiple items:
```bash
kroger cart add --items '[{"upc":"<upc1>","quantity":1},{"upc":"<upc2>","quantity":2}]' --json
```

## Supported Chains

This CLI works with all Kroger-owned stores: Kroger, Fred Meyer, Ralphs, Harris Teeter, King Soopers, Dillons, Fry's, QFC, Smith's, Jay C, Food 4 Less, FoodsCo, Mariano's, Metro Market, Pick 'n Save, City Market. Use `--chain` on store search to filter.

## Important Notes

- A default store must be selected for product searches (pricing is store-specific)
- The cart is **add-only** - you cannot view, remove, or modify items via the API
- You **cannot** place orders, schedule pickup/delivery, or checkout via the API
- After adding items, tell the user to go to their store's website (e.g. fredmeyer.com, kroger.com) to review their cart, schedule pickup or delivery, and complete checkout
- Each product in the JSON output includes a `url` field linking directly to the product page
- All commands output JSON automatically when piped (non-TTY), or with `--json`
- Use `--quiet` to suppress status messages

## Handling $ARGUMENTS

When the user says something like "add milk and eggs to my Kroger cart":
1. Search for each item separately
2. Present options with prices to the user
3. Confirm before adding to cart
4. Add confirmed items in a single batch call
