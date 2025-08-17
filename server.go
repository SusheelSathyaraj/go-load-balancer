package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
)

// Backend server struct
type Server struct {
	Address   string
	IsHealthy bool
	Mutex     sync.Mutex
	ConCount  int //for least connection algo
	URL       *url.URL
}

// creating a new server instance
func NewServer(address string) (*Server, error) {
	serverURL, err := url.Parse(address)

	if err != nil {
		return nil, fmt.Errorf("invalid server URL %s: %v", address, err)
	}

	return &Server{
		Address:   address,
		IsHealthy: false, //will be set by the health chech methods
		ConCount:  0,
		URL:       serverURL,
	}, nil
}

// Active connections count

func (s *Server) ActiveConnectionsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		s.Mutex.Lock()
		defer s.Mutex.Unlock()

		healthStatus := "unhealthy"
		if s.IsHealthy {
			healthStatus = "healthy"
		}

		log.Printf("The number of active connections on port %s are %d and the server is %s", s.Address, s.ConCount, healthStatus)
		fmt.Fprintf(w, "The number of active connections on port %s are %d and the server is %s", s.Address, s.ConCount, healthStatus)
	}
}
