package httpserver

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Gurveer1510/logaggregator/internal/buffer"
	"github.com/Gurveer1510/logaggregator/internal/hub"
	"github.com/Gurveer1510/logaggregator/internal/model"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Server struct {
	ringBuffer *buffer.RingBuffer
	port       string
	h          *hub.Hub
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewServer(ringBuffer *buffer.RingBuffer, port string, hub *hub.Hub) *Server {
	return &Server{
		ringBuffer: ringBuffer,
		h:          hub,
		port:       port,
	}
}

func (s *Server) Start() {
	router := mux.NewRouter()
	router.HandleFunc("/log", s.logHandler).Methods("GET")
	router.HandleFunc("/ws", s.handleWs)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend"))).Methods("GET")

	log.Println("Server started on http://localhost:" + s.port)

	if err := http.ListenAndServe(":"+s.port, router); err != nil {
		panic(err)
	}
}

func (s *Server) handleWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	ch := make(chan model.LogEntry, 10)
	s.h.Subscribe(ch)
	defer func(){
		s.h.Unsubscribe(ch)
		close(ch)
	}()

	for v := range ch {
		if err := conn.WriteJSON(v); err != nil {
			break
		}
	}
}

func (s *Server) logHandler(w http.ResponseWriter, r *http.Request) {
	level := r.URL.Query().Get("level")
	logs := s.ringBuffer.Query()

	output := []*model.LogEntry{}

	if level == "" {
		output = logs
	} else {
		for _, log := range logs {
			if log.Level == model.LogLevel(level) {
				output = append(output, log)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(output)
}
