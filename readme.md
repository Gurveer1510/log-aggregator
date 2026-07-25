# Log Aggregator

A real-time log aggregation service built in Go. Accepts structured logs over TCP, stores them in a ring buffer, and pushes them to connected clients via WebSocket.

## Architecture

```
┌────────────┐   TCP (JSON lines)   ┌──────────────┐   broadcast   ┌──────────────┐
│ Log Source  │ ──────────────────── │  TCP Server  │ ────────────► │     Hub      │
│ (any app)  │    :8001             │              │               │  (fan-out)   │
└────────────┘                      └──────┬───────┘               └──────┬───────┘
                                           │ insert                       │
                                    ┌──────▼───────┐               ┌──────▼───────┐
                                    │ Ring Buffer   │               │  WebSocket   │
                                    │ (fixed-size)  │               │  Clients     │
                                    └──────┬───────┘               └──────────────┘
                                           │ query
                                    ┌──────▼───────┐
                                    │  HTTP API    │
                                    │  :8000       │
                                    └──────────────┘
```

## Quick Start

```bash
# clone and run
git clone https://github.com/Gurveer1510/log-aggregator.git
cd log-aggregator

# create a .env file
echo "HTTP_PORT=8000
TCP_PORT=8001
BUFFER_SIZE=100" > .env

go run cmd/main.go
```

## Sending Logs

Send newline-delimited JSON to the TCP port:

```bash
echo '{"level":"INFO","timestamp":"2025-07-26 10:00:00","service_name":"auth","message":"user logged in"}' | nc localhost 8001
echo '{"level":"ERROR","timestamp":"2025-07-26 10:00:01","service_name":"payments","message":"charge failed"}' | nc localhost 8001
```

Each log entry has four fields:

| Field          | Type   | Values                    |
|----------------|--------|---------------------------|
| `level`        | string | `INFO`, `ERROR`, `DEBUG`  |
| `timestamp`    | string | any format                |
| `service_name` | string | name of the source service|
| `message`      | string | log message               |

## Querying Logs

```bash
# get all buffered logs
curl http://localhost:8000/log

# filter by level
curl http://localhost:8000/log?level=ERROR
```

## Live Dashboard

Open `http://localhost:8000` in a browser. Logs appear in real time via WebSocket. Use the dropdown to filter by level.

## Configuration

Set via `.env` file or environment variables:

| Variable      | Default | Description                  |
|---------------|---------|------------------------------|
| `HTTP_PORT`   | 8000    | HTTP server and dashboard    |
| `TCP_PORT`    | 8001    | TCP ingestion port           |
| `BUFFER_SIZE` | 100     | Ring buffer capacity (entries)|

## Project Structure

```
cmd/
  main.go                       # entry point
  server/
    httpserver/main.go          # HTTP API + WebSocket + dashboard
    tcpserver/main.go           # TCP log ingestion
internal/
  buffer/ring-buffer.go         # thread-safe ring buffer
  config/config.go              # env-based configuration
  hub/hub.go                    # WebSocket fan-out broadcaster
  model/model.go                # LogEntry struct
frontend/
  index.html                    # live dashboard
```