# zerodha-kite-cli

CLI-based tooling for Zerodha Kite workflows.

> **Note**: This is an unofficial implementation. It uses the official Go SDK [gokiteconnect](https://github.com/zerodha/gokiteconnect).

## Prerequisites (Build from Source)

- Go `1.26.0` (exact version)

## Install (Prebuilt Binary)

Install the latest release without cloning the repo or building from source.

Linux/macOS (curl):
```bash
curl -fsSL https://raw.githubusercontent.com/jatinbansal1998/zerodha-kite-cli/main/scripts/install.sh | sh
```

Linux/macOS (wget):
```bash
wget -qO- https://raw.githubusercontent.com/jatinbansal1998/zerodha-kite-cli/main/scripts/install.sh | sh
```

Linux/macOS (pin a specific version):
```bash
curl -fsSL https://raw.githubusercontent.com/jatinbansal1998/zerodha-kite-cli/main/scripts/install.sh | sh -s -- --version v1.2.3
```

Windows PowerShell:
```powershell
irm https://raw.githubusercontent.com/jatinbansal1998/zerodha-kite-cli/main/scripts/install.ps1 | iex
```

Windows CMD:
```bat
powershell -NoProfile -ExecutionPolicy Bypass -Command "irm https://raw.githubusercontent.com/jatinbansal1998/zerodha-kite-cli/main/scripts/install.ps1 | iex"
```

Windows PowerShell (pin a specific version):
```powershell
$env:ZERODHA_VERSION='v1.2.3'
irm https://raw.githubusercontent.com/jatinbansal1998/zerodha-kite-cli/main/scripts/install.ps1 | iex
```

## Toolchain

This repository pins the Go toolchain in `go.mod` via:

- `go 1.26`
- `toolchain go1.26.0`

## Build

```bash
go build ./cmd/zerodha
```

## Manual Update

Run an on-demand self-update check and apply the latest release:

```bash
zerodha update
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
zerodha profile full
zerodha quote get NSE:INFY NSE:TCS
zerodha quote ltp NSE:INFY NSE:TCS
zerodha quote ohlc NSE:INFY NSE:TCS
zerodha quote historical --instrument-token 408065 --interval day --from 2026-01-01 --to 2026-02-01
zerodha instruments list
zerodha instruments list --exchange NSE
zerodha instruments mf
zerodha gtt list
zerodha mf orders list
zerodha mf sips list
zerodha mf holdings
zerodha orders list
zerodha orders trades
zerodha orders trades --order-id <order_id>
zerodha positions
zerodha positions convert --exchange NSE --symbol INFY --old-product CNC --new-product MIS --position-type day --txn BUY --qty 1
zerodha holdings
zerodha holdings auctions
zerodha holdings auth-initiate --type equity --transfer-type pre
zerodha margins --segment all
zerodha margins order --exchange NSE --symbol INFY --txn BUY --type MARKET --product CNC --qty 1
zerodha margins basket --exchange NSE --symbol INFY --txn BUY --type MARKET --product CNC --qty 1 --consider-positions
zerodha margins charges --exchange NSE --symbol INFY --txn BUY --type MARKET --product CNC --qty 1 --avg-price 1500
```
4. Place order:
```bash
zerodha order place --exchange NSE --symbol INFY --txn BUY --type MARKET --product CNC --qty 1
zerodha order exit --order-id <order_id> --variety regular
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

Other session commands:

- `zerodha auth renew` renews access token using stored refresh token.
- `zerodha auth logout` invalidates access token (if possible) and clears local tokens.
- `zerodha auth revoke-refresh [--refresh-token <token>]` invalidates refresh token (defaults to stored token).

## SDK Coverage Tracking

- Coverage and implementation plan are tracked in `SDK_COVERAGE_ROADMAP.md`.
