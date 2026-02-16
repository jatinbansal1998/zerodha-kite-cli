# zerodha-kite-cli

CLI-based tooling for Zerodha Kite workflows.

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
2. Login:
```bash
zerodha auth login
```
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

## Modes

- Interactive mode: `zerodha -i`
- Non-interactive mode: `zerodha <subcommand>`
