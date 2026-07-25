package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HttpPort string `env:"HTTP_PORT" envDefault:"8000"`
	TcpPort  string `env:"TCP_PORT" envDefault:"8001"`

	BufferSize int `env:"BUFFER_SIZE" envDefault:"100"`
}

func New() *Config {
	return &Config{}
}

func (c *Config) LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Read variables using the standard os package
	httpPort := os.Getenv("HTTP_PORT")
	tcpPort := os.Getenv("TCP_PORT")
	bufferSize := os.Getenv("BUFFER_SIZE")

	if httpPort != "" {
		c.HttpPort = httpPort
	} else {
		c.HttpPort = "8000" // Default value
	}

	if tcpPort != "" {
		c.TcpPort = tcpPort
	} else {
		c.TcpPort = "8001" // Default value
	}

	if bufferSize != "" {
		fmt.Sscanf(bufferSize, "%d", &c.BufferSize)
	} else {
		c.BufferSize = 100 // Default value
	}
}
