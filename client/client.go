package client

import (
	"encoding/json"
	"net"
	"sync"
	"time"
)

type LogLevel string

const (
	ErrorLevel LogLevel = "ERROR"
	DebugLevel LogLevel = "DEBUG"
	InfoLevel  LogLevel = "INFO"
)

type LogEntry struct {
	Level       LogLevel        `json:"level"`
	Timestamp   string          `json:"timestamp"`
	ServiceName string          `json:"service_name"`
	Message     json.RawMessage `json:"message"`
}

type Client struct {
	addr string
	conn net.Conn
	mu   sync.Mutex
}

func NewClient(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Client{
		addr: addr,
		conn: conn,
	}, nil
}

func (c *Client) log(level LogLevel, service string, message any) error {
	parsedMessage, err := parseMessage(message)
	if err != nil {
		return err
	}

	entry := LogEntry{
		Level:       level,
		Timestamp:   time.Now().Format(time.RFC3339),
		ServiceName: service,
		Message:     parsedMessage,
	}

	return c.writeConn(entry)
}

func (c *Client) Info(service string, message any) error {
    return c.log(InfoLevel, service, message)
}
func (c *Client) Debug(service string, message any) error {
    return c.log(DebugLevel, service, message)
}
func (c *Client) Error(service string, message any) error {
    return c.log(ErrorLevel, service, message)
}

func (c *Client) writeConn(entry LogEntry) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	data = append(data, '\n') // Add a newline to separate log entries

	_, err = c.conn.Write(data)

	if err != nil {
		return err
	}

	return nil
}

func parseMessage(message any) (json.RawMessage, error) {
	jsonMsg, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	return jsonMsg, nil

}

func (c *Client) Close() error {
    return c.conn.Close()
}