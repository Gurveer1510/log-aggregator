package main

import (
	httpserver "github.com/Gurveer1510/logaggregator/cmd/server/httpserver"
	"github.com/Gurveer1510/logaggregator/cmd/server/tcpserver"
	"github.com/Gurveer1510/logaggregator/internal/buffer"
	"github.com/Gurveer1510/logaggregator/internal/hub"
	"github.com/Gurveer1510/logaggregator/internal/model"
)

func main() {
	ringBuffer := buffer.NewRingBuffer(10)
	hub := hub.NewHub()
	tcpServer := tcpserver.NewServer(ringBuffer, hub)
	httpServer := httpserver.NewServer(ringBuffer, hub)


	ch1 := make(chan model.LogEntry, 10)
	ch2 := make(chan model.LogEntry, 10)

	hub.Subscribe(ch1)
	hub.Subscribe(ch2)

	go tcpServer.Start()
	httpServer.Start()
}
