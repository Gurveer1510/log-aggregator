# Log Aggregator

A real-time log aggregation service built in Go. Accepts structured logs over TCP, stores them in a thread-safe ring buffer, and pushes them to connected clients via WebSocket.

## Architecture

```
┌────────────┐   TCP (JSON lines)   ┌──────────────┐   broadcast   ┌──────────────┐
│ Log Source  │ ──────────────────── │  TCP Server  │ ────────────► │     Hub      │
│ (any app)  │                      │              │               │  (fan-out)   │
└────────────┘                      └──────┬───────┘               └──────┬───────┘
                                           │ insert                       │
                                    ┌──────▼───────┐               ┌──────▼───────┐
                                    │ Ring Buffer   │               │  WebSocket   │
                                    │ (fixed-size)  │               │  Clients     │
                                    └──────┬───────┘               └──────────────┘
                                           │ query
                                    ┌──────▼───────┐
                                    │  HTTP API    │
                                    │  + Dashboard │
                                    └──────────────┘
```

## Quick Start

```bash
git clone https://github.com/Gurveer1510/log-aggregator.git
cd log-aggregator

# create a .env file
echo "HTTP_PORT=8000
TCP_PORT=8001
BUFFER_SIZE=100" > .env

go run cmd/main.go
```

## Client SDK

Other Go projects can send logs using the built-in client package:

```bash
go get github.com/Gurveer1510/log-aggregator
```

```go
package main

import (
    "log"
    "github.com/Gurveer1510/log-aggregator/pkg/client"
)

func main() {
    logger, err := client.NewClient("localhost:8001")
    if err != nil {
        log.Fatal(err)
    }
    defer logger.Close()

    // simple string messages
    logger.Info("auth", "user logged in")
    logger.Error("payments", "charge failed")

    // structured data
    logger.Info("auth", map[string]any{
        "user_id": 123,
        "action":  "login",
        "ip":      "192.168.1.1",
    })
}
```

The client maintains a persistent TCP connection, handles concurrent writes via mutex, and supports string or structured JSON messages.

## Sending Logs Manually

You can also send logs directly via TCP without the SDK:

```bash
echo '{"level":"INFO","timestamp":"2025-07-26 10:00:00","service_name":"auth","message":"user logged in"}' | nc localhost 8001
echo '{"level":"ERROR","timestamp":"2025-07-26 10:00:01","service_name":"payments","message":"charge failed"}' | nc localhost 8001
```

Each log entry has four fields:

| Field          | Type              | Values                   |
|----------------|-------------------|--------------------------|
| `level`        | string            | `INFO`, `ERROR`, `DEBUG` |
| `timestamp`    | string            | any format               |
| `service_name` | string            | source service name      |
| `message`      | string or object  | log message or JSON data |

## Querying Logs

```bash
# get all buffered logs
curl http://localhost:8000/log

# filter by level
curl http://localhost:8000/log?level=ERROR
```

## Live Dashboard

Open `http://localhost:8000` in a browser. Logs appear in real time via WebSocket with level filtering and a connection status indicator.

## Configuration

Set via `.env` file or environment variables:

| Variable      | Default | Description                   |
|---------------|---------|-------------------------------|
| `HTTP_PORT`   | 8000    | HTTP API and dashboard        |
| `TCP_PORT`    | 8001    | TCP ingestion port            |
| `BUFFER_SIZE` | 100     | Ring buffer capacity (entries)|

## Project Structure

```
cmd/
  main.go                          # entry point
  server/
    httpserver/main.go             # HTTP API + WebSocket + dashboard
    tcpserver/main.go              # TCP log ingestion
internal/
  buffer/ring-buffer.go            # thread-safe ring buffer
  config/config.go                 # env-based configuration
  hub/hub.go                       # WebSocket fan-out broadcaster
  model/model.go                   # LogEntry struct
pkg/
  client/client.go                 # client SDK for sending logs
frontend/
  index.html                       # live dashboard
```

## Key Concepts

- **Ring Buffer** — fixed-size circular buffer that overwrites oldest entries when full, protected by `sync.RWMutex` for concurrent reads and writes
- **TCP Ingestion** — goroutine-per-connection model, parses newline-delimited JSON, supports multiple concurrent producers
- **Fan-out Hub** — maintains a subscriber registry of Go channels, broadcasts new entries to all connected WebSocket clients with non-blocking sends
- **WebSocket Streaming** — real-time push to browser clients with automatic cleanup on disconnect