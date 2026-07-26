# Log Aggregator — Usage Guide

A real-time log aggregation service in Go. Run it as a standalone service, send logs from any Go application using the client SDK, and monitor everything through a live browser dashboard.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Setting Up the Aggregator](#setting-up-the-aggregator)
- [Client SDK](#client-sdk)
- [Sending Logs Manually](#sending-logs-manually)
- [HTTP Query API](#http-query-api)
- [Live Dashboard](#live-dashboard)
- [Configuration Reference](#configuration-reference)
- [Project Structure](#project-structure)
- [Troubleshooting](#troubleshooting)

---

## Architecture Overview

The log aggregator runs as a standalone service, separate from your application. Your apps send logs to it via TCP, and you view them through a browser dashboard or HTTP API.

```
┌──────────────┐                       ┌───────────────────────────────────┐
│  App A       │──── TCP :8001 ──────► │         LOG AGGREGATOR           │
│  (Go SDK)    │                       │                                  │
└──────────────┘                       │  ┌───────────┐   ┌───────────┐  │
                                       │  │TCP Server │──►│Ring Buffer│  │
┌──────────────┐                       │  └─────┬─────┘   └─────┬─────┘  │
│  App B       │──── TCP :8001 ──────► │        │               │        │
│  (Go SDK)    │                       │        ▼               ▼        │
└──────────────┘                       │  ┌───────────┐   ┌───────────┐  │
                                       │  │   Hub     │   │ HTTP API  │  │
┌──────────────┐                       │  │ (fan-out) │   │ /log      │  │
│  App C       │──── TCP :8001 ──────► │  └─────┬─────┘   └───────────┘  │
│  (nc/curl)   │                       │        │                        │
└──────────────┘                       │        ▼                        │
                                       │  ┌───────────┐                  │
┌──────────────┐                       │  │ WebSocket │◄── Browser       │
│  Browser     │◄─── HTTP :8000 ──────►│  │ /ws       │   Dashboard      │
│  Dashboard   │                       │  └───────────┘                  │
└──────────────┘                       └───────────────────────────────────┘
```

The aggregator has four core components:

- **TCP Server** — accepts persistent connections and reads newline-delimited JSON logs, one entry per line, using a goroutine-per-connection model.
- **Ring Buffer** — a fixed-size circular buffer protected by a read-write mutex. When full, the oldest entries are overwritten. This means the aggregator always holds the most recent N log entries in memory.
- **Hub** — a fan-out broadcaster. When a new log arrives, the hub pushes it to every connected WebSocket client through Go channels. Slow clients are skipped (non-blocking send) so one slow browser tab never blocks other subscribers.
- **HTTP Server** — serves the query API (`/log`), the WebSocket endpoint (`/ws`), and the browser dashboard.

---

## Setting Up the Aggregator

### Prerequisites

- Go 1.21 or later

### Installation

```bash
git clone https://github.com/Gurveer1510/log-aggregator.git
cd log-aggregator
```

### Configuration

Create a `.env` file in the project root:

```env
HTTP_PORT=8000
TCP_PORT=8001
BUFFER_SIZE=100
```

All values are optional — defaults are shown above. If no `.env` file exists, the aggregator uses these defaults.

### Running

```bash
go run cmd/main.go
```

You should see:

```
Server is listening on port 8001...
Server started on http://localhost:8000
```

The aggregator is now ready to receive logs on TCP port 8001 and serve the dashboard on HTTP port 8000.

---

## Client SDK

The client SDK lets any Go application send structured logs to the aggregator with a single function call.

### Installation

In your Go project:

```bash
go get github.com/Gurveer1510/log-aggregator@main
```

If you hit a module cache issue, bypass the Go proxy:

```bash
GOPROXY=direct go get github.com/Gurveer1510/log-aggregator@main
```

### Basic Usage

```go
package main

import (
    "log"
    "github.com/Gurveer1510/log-aggregator/pkg/client"
)

func main() {
    // Connect to the aggregator
    logger, err := client.NewClient("localhost:8001")
    if err != nil {
        log.Fatal(err)
    }
    defer logger.Close()

    // Send logs
    logger.Info("auth", "user logged in")
    logger.Error("payments", "charge failed")
    logger.Debug("cache", "key expired")
}
```

### Structured Messages

The message field accepts any type — strings, maps, structs, or any value that can be JSON-marshalled:

```go
// Plain string
logger.Info("auth", "user logged in")

// Map with structured data
logger.Info("auth", map[string]any{
    "user_id": 123,
    "action":  "login",
    "ip":      "192.168.1.1",
})

// Custom struct
type OrderEvent struct {
    OrderID string  `json:"order_id"`
    Amount  float64 `json:"amount"`
    Status  string  `json:"status"`
}

logger.Info("orders", OrderEvent{
    OrderID: "ORD-456",
    Amount:  99.99,
    Status:  "completed",
})
```

### Available Methods

| Method | Level | Usage |
|--------|-------|-------|
| `logger.Info(service, message)` | INFO | General operational events |
| `logger.Error(service, message)` | ERROR | Failures and exceptions |
| `logger.Debug(service, message)` | DEBUG | Detailed diagnostic info |

Every method takes two arguments: a service name (string identifying the source) and a message (any type). The client automatically adds a timestamp in RFC3339 format.

### Thread Safety

The client is safe to use from multiple goroutines. All writes are protected by a mutex, so you can share a single client instance across your application:

```go
var logger *client.Client

func init() {
    var err error
    logger, err = client.NewClient("localhost:8001")
    if err != nil {
        log.Fatal(err)
    }
}

func HandleLogin(userID int) {
    // Safe to call from any goroutine
    logger.Info("auth", map[string]any{"user_id": userID, "action": "login"})
}

func HandlePayment(orderID string) {
    // Same client, different goroutine
    logger.Info("payments", map[string]any{"order_id": orderID})
}
```

### Connection Lifecycle

The client opens a persistent TCP connection when created and reuses it for all log calls. Always close it when your application shuts down:

```go
logger, err := client.NewClient("localhost:8001")
if err != nil {
    log.Fatal(err)
}
defer logger.Close()  // closes the TCP connection on exit
```

---

## Sending Logs Manually

You can send logs without the SDK using any tool that writes to a TCP socket. Each log entry is a single line of JSON followed by a newline character.

### Using netcat

```bash
echo '{"level":"INFO","timestamp":"2025-07-26 10:00:00","service_name":"auth","message":"user logged in"}' | nc localhost 8001
```

### Multiple entries in one session

```bash
nc localhost 8001 << 'EOF'
{"level":"INFO","timestamp":"2025-07-26 10:00:00","service_name":"auth","message":"user logged in"}
{"level":"ERROR","timestamp":"2025-07-26 10:00:01","service_name":"payments","message":"charge failed"}
{"level":"DEBUG","timestamp":"2025-07-26 10:00:02","service_name":"cache","message":"key expired"}
EOF
```

### Interactive session

```bash
nc localhost 8001
```

Then type JSON lines one at a time. Each line appears in the dashboard as soon as you press Enter.

### Log Entry Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `level` | string | yes | One of `INFO`, `ERROR`, `DEBUG` |
| `timestamp` | string | yes | Any format (displayed as-is) |
| `service_name` | string | yes | Identifies the source service |
| `message` | string or object | yes | Log content — plain text or structured JSON |

### Example Entries

```json
{"level":"INFO","timestamp":"2025-07-26T10:00:00Z","service_name":"api","message":"server started on port 3000"}

{"level":"ERROR","timestamp":"2025-07-26T10:00:01Z","service_name":"db","message":"connection refused"}

{"level":"INFO","timestamp":"2025-07-26T10:00:02Z","service_name":"auth","message":{"user_id":42,"action":"login","ip":"10.0.0.1"}}
```

---

## HTTP Query API

### GET /log

Returns all buffered log entries as a JSON array.

```bash
curl http://localhost:8000/log
```

Response:

```json
[
  {
    "level": "INFO",
    "timestamp": "2025-07-26T10:00:00Z",
    "service_name": "auth",
    "message": "user logged in"
  },
  {
    "level": "ERROR",
    "timestamp": "2025-07-26T10:00:01Z",
    "service_name": "payments",
    "message": "charge failed"
  }
]
```

### Filtering by Level

```bash
# Only errors
curl http://localhost:8000/log?level=ERROR

# Only info
curl http://localhost:8000/log?level=INFO

# Only debug
curl http://localhost:8000/log?level=DEBUG
```

### Notes

The API returns entries currently in the ring buffer. Once the buffer is full, the oldest entries are overwritten and no longer available. The number of entries retained is controlled by the `BUFFER_SIZE` configuration.

---

## Live Dashboard

Open `http://localhost:8000` in any browser.

### Features

- **Real-time streaming** — logs appear instantly as they arrive via WebSocket
- **Connection indicator** — green dot when connected, red when disconnected
- **Auto-reconnect** — reconnects automatically if the aggregator restarts
- **Level filtering** — dropdown to show only INFO, ERROR, or DEBUG entries
- **Entry counter** — shows total number of received entries
- **Clear button** — resets the displayed entries (does not affect the server buffer)

### WebSocket Endpoint

For custom integrations, connect directly to the WebSocket:

```javascript
const ws = new WebSocket("ws://localhost:8000/ws");

ws.onmessage = (event) => {
  const entry = JSON.parse(event.data);
  console.log(`[${entry.level}] ${entry.service_name}: ${entry.message}`);
};
```

Each message is a single JSON log entry. Entries are pushed as they arrive — there is no batching or buffering on the WebSocket side.

---

## Configuration Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_PORT` | `8000` | Port for the HTTP server (API, dashboard, WebSocket) |
| `TCP_PORT` | `8001` | Port for TCP log ingestion |
| `BUFFER_SIZE` | `100` | Number of log entries the ring buffer holds. When full, oldest entries are overwritten. |

Configuration is loaded from a `.env` file in the project root. If the file is missing, defaults are used. Environment variables override `.env` values.

---

## Project Structure

```
log-aggregator/
├── cmd/
│   ├── main.go                          # entry point — wires everything together
│   └── server/
│       ├── httpserver/main.go           # HTTP API, WebSocket handler, static file serving
│       └── tcpserver/main.go            # TCP listener, JSON line parsing
├── internal/
│   ├── buffer/ring-buffer.go            # thread-safe circular buffer
│   ├── config/config.go                 # .env and environment variable loading
│   ├── hub/hub.go                       # subscriber registry and fan-out broadcaster
│   └── model/model.go                   # LogEntry struct definition
├── pkg/
│   └── client/client.go                 # client SDK for sending logs
├── frontend/
│   └── index.html                       # browser dashboard
├── go.mod
├── go.sum
└── .env                                 # configuration (not committed)
```

### Package Responsibilities

| Package | Role |
|---------|------|
| `cmd/main.go` | Creates the ring buffer, hub, and both servers. Starts TCP in a goroutine, HTTP on the main goroutine. |
| `tcpserver` | Listens for TCP connections. Each connection gets a goroutine that scans lines, unmarshals JSON, inserts into the ring buffer, and broadcasts via the hub. |
| `httpserver` | Serves `GET /log` for querying, `/ws` for WebSocket streaming, and `/` for the dashboard. |
| `buffer` | Ring buffer with `Insert` and `Query`. Fixed size, overwrites oldest on overflow. `sync.RWMutex` for safe concurrent access. |
| `hub` | Maintains a slice of subscriber channels. `Subscribe` adds, `Unsubscribe` removes, `Broadcast` pushes to all with non-blocking sends. |
| `model` | Defines `LogEntry` with JSON tags. `Message` field uses `json.RawMessage` to accept both strings and objects. |
| `pkg/client` | TCP client with `Info`/`Error`/`Debug` methods. Handles JSON marshalling, newline framing, and mutex-protected writes. |

---

## Troubleshooting

### "connection refused" when using the client SDK

The aggregator isn't running, or it's on a different port. Check that `go run cmd/main.go` is running and that the TCP port matches what you passed to `NewClient`.

### Logs not appearing in the dashboard

Make sure the browser is connected (green dot in the header). If it's red, the WebSocket connection failed — check that the HTTP port is correct and the aggregator is running.

Verify your JSON tags match: the fields must be `level`, `timestamp`, `service_name`, and `message` (all lowercase).

### `go get` fails with module path mismatch

The Go module proxy may have cached an old version. Bypass it:

```bash
GOPROXY=direct go get github.com/Gurveer1510/log-aggregator/pkg/client@main
```

### `[object Object]` in the dashboard

The frontend needs to stringify structured messages. In `frontend/index.html`, update the `logRow` function to handle object messages:

```javascript
const message = typeof entry.message === "object"
    ? JSON.stringify(entry.message)
    : entry.message || "";
```

### Buffer fills up and old logs disappear

This is expected. The ring buffer has a fixed capacity set by `BUFFER_SIZE`. When it's full, the oldest entries are overwritten. Increase `BUFFER_SIZE` in your `.env` file to retain more entries.