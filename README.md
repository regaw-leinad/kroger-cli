# kroger-cli

A command-line interface for the Kroger API. Search for products, find stores, and add items to your cart - designed for AI agents and humans alike.

Works with all Kroger-owned stores: **Kroger**, **Fred Meyer**, **Ralphs**, **QFC**, **Dillons**, **Food 4 Less**, and [many more](#supported-chains).

## Quick Start

```bash
# Install
go install github.com/regaw-leinad/kroger-cli@latest

# Set up your API credentials (one-time)
kroger auth setup

# Log in with your Kroger account
kroger auth login

# Find and select your store
kroger store search --zip 98270
kroger store select 70100209

# Search for products
kroger product search "organic milk" --limit 5

# Add to cart
kroger cart add 0001111042850 --quantity 2
```

## Why?

AI agents can shop for you. Give an agent access to the `kroger` CLI and it can search products, compare prices, and build your cart. You just review and checkout.

The CLI outputs JSON automatically when used by agents (non-TTY), and renders human-friendly tables in the terminal.

## Installation

<details>
<summary><strong>Go install</strong></summary>

```bash
go install github.com/regaw-leinad/kroger-cli@latest
```

Requires Go 1.21+.
</details>

<details>
<summary><strong>Build from source</strong></summary>

```bash
git clone https://github.com/regaw-leinad/kroger-cli.git
cd kroger-cli
task build
# Binary is at bin/kroger
```

Requires [Task](https://taskfile.dev).
</details>

## Setup

You need a free Kroger Developer account to use this CLI. See the [setup guide](docs/setup.md) for step-by-step instructions.

```bash
# 1. Enter your API credentials
kroger auth setup

# 2. Log in via browser
kroger auth login

# 3. Find your store
kroger store search --zip 45202

# 4. Set your default store
kroger store select 01400413

# Check everything is working
kroger auth status
```

## Commands

### Products

```bash
# Search for products at your default store
kroger product search "sourdough bread"

# Filter by brand
kroger product search "milk" --brand "Simple Truth"

# Limit results
kroger product search "chicken" --limit 5

# Filter by fulfillment type
kroger product search "eggs" --fulfillment pickup

# Show product details
kroger product show 0001111042850
```

### Cart

```bash
# Add a single item
kroger cart add 0001111042850

# Add with quantity
kroger cart add 0001111042850 --quantity 3

# Add multiple items at once
kroger cart add --items '[{"upc":"0001111042850","quantity":2},{"upc":"0001111040190","quantity":1}]'
```

> **Note:** The Kroger API only supports adding items to your cart. To view your cart, remove items, or checkout, visit your store's website (e.g. fredmeyer.com, kroger.com).

### Stores

```bash
# Search by zip code
kroger store search --zip 98270

# Search by coordinates
kroger store search --lat 47.6 --lon -122.3

# Filter by chain
kroger store search --zip 98270 --chain "Fred Meyer"

# Set your default store
kroger store select 70100209

# View store details
kroger store show
```

### Auth

```bash
# Configure API credentials
kroger auth setup

# Log in via browser
kroger auth login

# Check status
kroger auth status

# Log out and clear credentials
kroger auth logout
```

## Agent Integration

The CLI is designed for AI agents. When stdout is not a terminal (piped or run by an agent), output is automatically JSON.

```bash
# Agent gets JSON automatically
kroger product search "avocado" --limit 3
```

```json
[{"upc":"0000000004046","brand":"Fresh Avocados","description":"Fresh Medium Ripe Avocado","size":"1 each","price":1,"url":"https://www.fredmeyer.com/p/fresh-medium-ripe-avocado/0000000004046",...}]
```

Every product includes a `url` field linking directly to the product page and an `image` field with a product photo.

### Claude Code Plugin

Install as a Claude Code plugin to get the `/kroger` skill:

```bash
claude plugin install github.com/regaw-leinad/kroger-cli
```

Then just ask Claude to shop for you.

## Output Modes

| Context | Format | Example |
|---------|--------|---------|
| Terminal (human) | Aligned table | `kroger product search "milk"` |
| Non-TTY (agent) | Compact JSON | `kroger product search "milk" \| jq .` |
| Explicit JSON | Compact JSON | `kroger product search "milk" --json` |
| Quiet | Suppress status messages | `kroger product search "milk" --quiet` |

## Supported Chains

This CLI works with all Kroger-owned stores:

| Chain | Website |
|-------|---------|
| Kroger | kroger.com |
| Fred Meyer | fredmeyer.com |
| Ralphs | ralphs.com |
| Harris Teeter | harristeeter.com |
| King Soopers | kingsoopers.com |
| Dillons | dillons.com |
| Fry's | frysfood.com |
| QFC | qfc.com |
| Smith's | smithsfoodanddrug.com |
| Food 4 Less | food4less.com |
| FoodsCo | foodsco.net |
| Mariano's | marianos.com |
| Metro Market | metromarket.net |
| Pick 'n Save | picknsave.com |
| City Market | citymarket.com |
| Jay C | jaycfoods.com |

## Configuration

Credentials are stored in `~/.config/kroger/credentials.json` (file permissions `0600`).

Non-sensitive config (default store) is stored in `~/.config/kroger/config.json`.

Environment variables override stored credentials:

| Variable | Purpose |
|----------|---------|
| `KROGER_CLIENT_ID` | Override client ID |
| `KROGER_CLIENT_SECRET` | Override client secret |
| `KROGER_CONFIG_DIR` | Override config directory |

## License

MIT
