package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	HttpPort string `env:"HTTP_PORT"`
	TcpPort  string `env:"TCP_PORT"`

	BufferSize int `env:"BUFFER_SIZE"`
}

func New() *Config {
	return &Config{}
}

func (c *Config) LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
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

	bufferSizeInt, err := strconv.Atoi(bufferSize)
	if err == nil && bufferSize != "" {
		c.BufferSize = bufferSizeInt
	} else {
		c.BufferSize = 100 // Default value
	}
}
