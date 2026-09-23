# Home Assistant Parts Inventory

A small standalone Go web application for recording and browsing a personal electronics-parts inventory. It can run directly on any machine or as a Home Assistant add-on.

## Run locally

```text
go run ./cmd/parts-inventory -db ./data/parts.db -addr :8080
```

Open `http://localhost:8080`. The database directory is created automatically. For a production or Home Assistant deployment, the default database path is `/data/parts.db`.

## Test and build

```text
go test ./...
go build ./cmd/parts-inventory
```

## Home Assistant add-on

The repository root is an Ingress-enabled Home Assistant app package, while the Go code remains a normal standalone application. Add this repository to Home Assistant's app store, install **Parts Inventory**, start it, and open the Web UI. The app runs the normal Go application and persists inventory in `/data/parts.db`.
