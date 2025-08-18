package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

//Test Helper function

// creates a  test HTTP server with customisable responses
func createMockServer(response string, statusCode int, delay time.Duration) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if delay > 0 {
			time.Sleep(delay)
		}
		if r.URL.Path == "/health" {
			if statusCode == http.StatusOK {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("healthy"))
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("unhealthy"))
			}
			return
		}
		w.WriteHeader(statusCode)
		w.Write([]byte(response))
	}))
}

// creates a set of test servers
func createTestServers(count int, healthy bool) ([]*Server, []*httptest.Server) {
	servers := make([]*Server, count)
	testServers := make([]*httptest.Server, count)

	for i := 0; i < count; i++ {
		statusCode := http.StatusOK
		if !healthy {
			statusCode = http.StatusServiceUnavailable
		}

		testServer := createMockServer(fmt.Sprintf("server %d", i+1), statusCode, 0)
		testServers[i] = testServer

		serverURL, _ := url.Parse(testServer.URL)
		servers[i] = &Server{
			Address:   testServer.URL,
			IsHealthy: healthy,
			URL:       serverURL,
		}
	}
	return servers, testServers
}

// cleanup closes all test servers
func cleanup(testServers []*httptest.Server) {
	for _, server := range testServers {
		server.Close()
	}
}

// Test Server struct
func TestNewServer(t *testing.T) {
	t.Run("Valid URL", func(t *testing.T) {
		server, err := NewServer("http://localhost:8081")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if server.Address != "http://localhost:8081" {
			t.Errorf("Expected address http://localhost:8081, got %s", server.Address)
		}

		if server.IsHealthy {
			t.Errorf("Expected server to be unhealthy initially")
		}

		if server.ConCount != 0 {
			t.Errorf("Expected concount to be 0, got %d", server.ConCount)
		}
	})

	t.Run("Invalid URL", func(t *testing.T) {
		_, err := NewServer("invalid-url")
		if err != nil {
			t.Errorf("Expected error for invalid URL")
		}
	})
}

func TestServerConnectionManagement(t *testing.T) {
	server, _ := NewServer("http://localhost:8081")

	//test initial state
	if server.GetConnectionCount() != 0 {
		t.Errorf("Expected the initial connection count to be 0, got %d", server.GetConnectionCount())
	}

	//test increment
	server.IncrementConnectionCount()
	if server.GetConnectionCount() != 1 {
		t.Errorf("Expected connection count to be 1, got %d", server.GetConnectionCount())
	}

	//test multiple increments
	server.IncrementConnectionCount()
	server.IncrementConnectionCount()
	if server.GetConnectionCount() != 3 {
		t.Errorf("Expected connection count to be 3, got %d", server.GetConnectionCount())
	}

	//test decrement
	server.DecrementConnectionCount()
	if server.GetConnectionCount() != 2 {
		t.Errorf("Expected connection count to be 2, got %d", server.GetConnectionCount())
	}

	//test multiple decrements
	server.DecrementConnectionCount()
	server.DecrementConnectionCount()
	if server.GetConnectionCount() != 0 {
		t.Errorf("Expected connection count to be 0, got %d", server.GetConnectionCount())
	}

	//test decrement below 0, should stay at 0
	server.DecrementConnectionCount()
	if server.GetConnectionCount() != 0 {
		t.Errorf("Expected connection count to remain at 0, got %d", server.GetConnectionCount())
	}
}

// Test for Load Balance
func TestNewLoadBalancer(t *testing.T) {
	servers, testServers := createTestServers(3, true)
	defer cleanup(testServers)

	lb := NewLoadBalancer(servers, "round-robin")

	if lb == nil {
		t.Fatal("Expected Load Balaner to be created")
	}

	if len(lb.Servers) != 3 {
		t.Errorf("Expected 3 servers, got %d", len(lb.Servers))
	}

	if lb.Algo != "round-robin" {
		t.Errorf("Expected round-robin algorithm, got %s", lb.Algo)
	}

	if lb.GetServerCount() != 3 {
		t.Errorf("Expected server count of 3, got %d", lb.GetServerCount())
	}
}
