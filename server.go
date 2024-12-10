package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

// Backend server struct
type Server struct {
	Address   string
	IsHealthy bool
	Mutex     sync.Mutex
	ConCount  int //for least connection algo
}

// Active connections count

func (s *Server) activeConnectionsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		s.Mutex.Lock()
		defer s.Mutex.Unlock()

		log.Printf("The number of active connections on port %s are %d", s.Address, s.ConCount)
		fmt.Fprintf(w, "The number of active connections on port %s are %d", s.Address, s.ConCount)
	}
}
