package tcpserver

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/Gurveer1510/logaggregator/internal/buffer"
	"github.com/Gurveer1510/logaggregator/internal/hub"
	"github.com/Gurveer1510/logaggregator/internal/model"
)

type Server struct {
	ringBuffer *buffer.RingBuffer
	h          *hub.Hub
}

func NewServer(ringBuffer *buffer.RingBuffer, hub *hub.Hub) *Server {
	return &Server{
		ringBuffer: ringBuffer,
		h:          hub,
	}
}

func (s *Server) Start() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Printf("Error while starting the server: %v", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server is listening on port 8080...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connections: %v", err)
			continue
		}

		go s.handleClient(conn)
	}
}

func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		inputLog := model.LogEntry{}
		if err := scanner.Err(); err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("Connection closed")
				return
			}
			fmt.Printf("Error while scanning lines: %s\n", err)
			return
		}

		err := json.Unmarshal(scanner.Bytes(), &inputLog)
		if err != nil {
			fmt.Printf("ERROR while unmarshaling: %s\n", err)
			continue
		}
		s.ringBuffer.Insert(inputLog)
		s.h.Broadcast(inputLog)
	}
}
