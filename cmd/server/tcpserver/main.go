package tcpserver

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"

	"github.com/Gurveer1510/log-aggregator/internal/buffer"
	"github.com/Gurveer1510/log-aggregator/internal/hub"
	"github.com/Gurveer1510/log-aggregator/internal/model"
)

type Server struct {
	ringBuffer *buffer.RingBuffer
	port       string
	h          *hub.Hub
}

func NewServer(ringBuffer *buffer.RingBuffer, port string, hub *hub.Hub) *Server {
	return &Server{
		ringBuffer: ringBuffer,
		h:          hub,
		port:       port,
	}
}

func (s *Server) Start() {
	listener, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		log.Printf("Error while starting the server: %v", err)
		return
	}
	defer listener.Close()
	log.Println("Server is listening on port " + s.port + "...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connections: %v", err)
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

		err := json.Unmarshal(scanner.Bytes(), &inputLog)
		if err != nil {
			log.Printf("ERROR while unmarshaling: %s\n", err)
			continue
		}
		s.ringBuffer.Insert(inputLog)
		s.h.Broadcast(inputLog)
	}
	if err := scanner.Err(); err != nil {
		if errors.Is(err, io.EOF) {
			log.Println("Connection closed")
			return
		}
		log.Printf("Error while scanning lines: %s\n", err)
		return
	}
}
