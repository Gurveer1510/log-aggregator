package main

import (
	httpserver "github.com/Gurveer1510/logaggregator/cmd/server/httpserver"
	"github.com/Gurveer1510/logaggregator/cmd/server/tcpserver"
	"github.com/Gurveer1510/logaggregator/internal/buffer"
	"github.com/Gurveer1510/logaggregator/internal/config"
	"github.com/Gurveer1510/logaggregator/internal/hub"
)

func main() {
	config := config.New()
	config.LoadConfig()
	ringBuffer := buffer.NewRingBuffer(config.BufferSize)
	hub := hub.NewHub()
	tcpServer := tcpserver.NewServer(ringBuffer, config.TcpPort, hub)
	httpServer := httpserver.NewServer(ringBuffer, config.HttpPort, hub)

	go tcpServer.Start()
	httpServer.Start()
}
