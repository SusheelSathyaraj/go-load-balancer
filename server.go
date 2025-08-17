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
	Mutex     sync.RWMutex
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

// getting current connection count
func (s *Server) GetConnectionCount() int {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return s.ConCount
}

// Increments the connection count
func (s *Server) IncrementConnectionCount() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.ConCount++
}

// Decrements the connection count
func (s *Server) DecrementConnectionCount() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	if s.ConCount > 0 {
		s.ConCount--
	}
}

// Returning if the server is healthy (thread safe)
func (s *Server) IsServerHealthy() bool {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return s.IsHealthy
}

// setting the health status (thread safe)
func (s *Server) SetHealthy(healthy bool) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.IsHealthy = healthy
}

// returning server info as a map
func (s *Server) GetServerInfo() map[string]interface{} {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	return map[string]interface{}{
		"address":      s.Address,
		"healthy":      s.IsHealthy,
		"connnections": s.ConCount,
	}
}

// http handler for monitoring active connections
func (s *Server) ActiveConnectionsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		s.Mutex.RLock()
		defer s.Mutex.RUnlock()

		healthStatus := "unhealthy"
		if s.IsHealthy {
			healthStatus = "healthy"
		}

		response := fmt.Sprintf(`{
		"address":"%s",
		"active-connections":"%d",
		"status":"%s"
		}`, s.Address, s.ConCount, healthStatus)

		log.Printf("Status check for %s: %d connections, %s", s.Address, s.ConCount, healthStatus)
		fmt.Fprintf(w, response)
	}
}
