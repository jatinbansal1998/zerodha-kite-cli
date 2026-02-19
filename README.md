# zerodha-kite-cli

CLI-based tooling for Zerodha Kite workflows.

> **Note**: This is an unofficial implementation. It uses the official Go SDK [gokiteconnect](https://github.com/zerodha/gokiteconnect).

## Prerequisites

- Go `1.26.0` (exact version)

## Toolchain

This repository pins the Go toolchain in `go.mod` via:

- `go 1.26`
- `toolchain go1.26.0`

## Build

```bash
go build ./cmd/zerodha
```

## Config and Cache

- Config file: `~/.config/zerodha/config.json`
- Cache directory: OS-native cache root + `/zerodha` (via `os.UserCacheDir()`)

## Quick Start

1. Add a profile:
```bash
zerodha config profile add default --api-key <api_key> --api-secret <api_secret> --set-active
```
Optional: update credentials individually:
```bash
zerodha config profile set-api-key default --api-key <api_key>
zerodha config profile set-api-secret default --api-secret <api_secret>
```
2. Login:
```bash
zerodha auth login --request-token <request_token_or_redirect_url>
```
Manual mode and callback mode are mutually exclusive.

3. Fetch data:
```bash
zerodha profile show
zerodha quote get NSE:INFY NSE:TCS
zerodha orders list
zerodha positions
zerodha holdings
zerodha margins --segment all
```
4. Place order:
```bash
zerodha order place --exchange NSE --symbol INFY --txn BUY --type MARKET --product CNC --qty 1
```

## Profile Commands

- `zerodha config profile add <name> --api-key ... --api-secret ...` adds a new profile or updates an existing one.
- `zerodha config profile set-api-key <name> --api-key ...` updates only the API key.
- `zerodha config profile set-api-secret <name> --api-secret ...` updates only the API secret.

## Auth Login Modes

Exactly one login mode is required:

1. Manual token mode:
```bash
zerodha auth login --request-token <request_token_or_redirect_url>
```
2. Local callback mode:
```bash
zerodha auth login --callback
zerodha auth login --callback --callback-port 8787
```

Invalid combinations:

- `--request-token` and `--callback` together.
- `--callback-port` explicitly set without `--callback`.
