# Kroger Developer App Setup Guide

This guide walks you through creating a Kroger Developer account and registering your app so you can use the `kroger` CLI.

## Step 1: Create a Kroger Developer Account

1. Go to [developer.kroger.com](https://developer.kroger.com)
2. Click **Sign Up** and create an account
3. Verify your email address

## Step 2: Register a New Application

1. Log in to the [Kroger Developer Portal](https://developer.kroger.com)
2. Navigate to **Apps** in the top menu
3. Click **Create New App** (or similar)
4. Fill in the required fields:
   - **App Name**: Choose any name (must be globally unique, e.g., "YourName Personal Shopper")
   - **App Description**: Anything you like (e.g., "My personal grocery shopping CLI")
   - **Support Contact**: Your email address

## Step 3: Set the Redirect URI

This is critical for the login flow to work:

1. In your app's settings, find the **Redirect URI** field
2. Set it to exactly:
   ```
   http://127.0.0.1:8642/callback
   ```
3. Save your changes

## Step 4: Copy Your Credentials

After creating the app, you'll see:

- **Client ID** - a string like `yourappname-abc123`
- **Client Secret** - a long string (shown only once! Save it immediately)

**Important:** The client secret is only shown once. Copy it now or print the page to PDF.

## Step 5: Configure the CLI

Run the setup command and enter your credentials:

```bash
kroger auth setup
```

Paste your Client ID and Client Secret when prompted.

## Step 6: Log In

```bash
kroger auth login
```

This opens your browser. Log in with your regular Kroger account (the one you use for shopping at kroger.com).

## Step 7: Select a Store

```bash
kroger store search --zip 45202
kroger store select <store-id>
```

You're all set! Start searching for products and adding them to your cart.

## Scopes

Your app is automatically granted these API scopes:

| API | Scope | What it does |
|-----|-------|-------------|
| Products | `product.compact` | Search and view product details |
| Cart | `cart.basic:write` | Add items to your cart |
| Profile | `profile.compact` | View your profile info |
| Locations | (no scope needed) | Search for stores |

## Troubleshooting

### "redirect_uri mismatch" error during login
Make sure the redirect URI in your Kroger Developer app is set to exactly `http://127.0.0.1:8642/callback`.

### "invalid_client" error
Double-check your Client ID and Client Secret. Run `kroger auth setup` again if needed.

### Port 8642 is already in use
Another process is using port 8642. Close it and try again, or wait for it to free up.
