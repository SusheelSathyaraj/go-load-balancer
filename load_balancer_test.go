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

func TestServerHealthManagement(t *testing.T) {
	server, _ := NewServer("http://localhost:8081")

	//test initial state
	if server.IsHealthy {
		t.Errorf("Expected server initial state to be unhealthy")
	}

	//test setting healthy
	server.SetHealthy(true)
	if !server.IsHealthy {
		t.Errorf("Expected server state to be set to true")
	}

	//test setting unhealthy
	server.SetHealthy(false)
	if server.IsHealthy {
		t.Errorf("Expected server state to be set to false")
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

func TestRoundRobinBalancing(t *testing.T) {
	servers, testservers := createTestServers(3, true)
	defer cleanup(testservers)

	lb := NewLoadBalancer(servers, "round-robin")

	//testing basic round robin rotation
	firstServer := lb.GetNextServer()
	secondServer := lb.GetNextServer()
	thirdServer := lb.GetNextServer()
	fourthServer := lb.GetNextServer()

	if firstServer == nil || secondServer == nil || thirdServer == nil || fourthServer == nil {
		t.Errorf("Expected servers to be returned")
	}

	//cycle through servers
	if firstServer.Address == secondServer.Address {
		t.Errorf("Expected different servers in consecutive calls")
	}

	if secondServer.Address == thirdServer.Address {
		t.Errorf("Expected different servers in consecutive calls")
	}

	if firstServer.Address != fourthServer.Address {
		t.Errorf("Expected to cycle back to first server")
	}
}

func TestLeastConnectionBalancing(t *testing.T) {
	servers, testServers := createTestServers(3, true)
	defer cleanup(testServers)

	//set different connection counts
	servers[0].ConCount = 5
	servers[1].ConCount = 2
	servers[2].ConCount = 8

	lb := NewLoadBalancer(servers, "least-connections")

	selectServer := lb.GetNextServer()

	if selectServer == nil {
		t.Fatal("Expected server to be returned")
	}

	//should select server with least connections
	if selectServer.ConCount != 2 {
		t.Errorf("Expected server with 2 connections, as it has the least connections, got server with %d connections", selectServer.ConCount)
	}
}

func TestHealthyServerSelection(t *testing.T) {
	servers, testServers := createTestServers(3, true)
	defer cleanup(testServers)

	//make only the second server healhty
	servers[1].IsHealthy = true

	lb := NewLoadBalancer(servers, "round-robin")

	//should always return the healthy server
	for i := 0; i < 5; i++ {
		selectedServer := lb.GetNextServer()

		if selectedServer == nil {
			t.Fatal("Expected healthy server to be returned")
		}

		if !selectedServer.IsHealthy {
			t.Errorf("Expected to get a healthy server")
		}

		if selectedServer.Address != servers[1].Address {
			t.Errorf("Expected to get healthy server, got %s ", selectedServer.Address)
		}
	}
}

func TestNoHealthyServers(t *testing.T) {
	servers, testServers := createTestServers(3, true)
	defer cleanup(testServers)

	lb := NewLoadBalancer(servers, "round-robin")

	selectedServer := lb.GetNextServer()

	if selectedServer != nil {
		t.Error("Expected no server when all servers are unhealthy")
	}
}

func TestLoadBalancerServerManagement(t *testing.T) {
	servers, testServers := createTestServers(2, true)
	defer cleanup(testServers)

	lb := NewLoadBalancer(servers, "round-robin")

	//test initial count
	if lb.GetServerCount() != 2 {
		t.Errorf("Expected 2 servers, got %d", lb.GetServerCount())
	}

	//test adding servers
	newTestServer := createMockServer("new-server", http.StatusOK, 0)
	defer newTestServer.Close()

	newServerURL, _ := url.Parse(newTestServer.URL)
	newServer := &Server{
		Address:   newTestServer.URL,
		IsHealthy: true,
		URL:       newServerURL,
	}

	lb.AddServer(newServer)

	if lb.GetServerCount() != 3 {
		t.Errorf("Expected 3 servers after adding server, got %d", lb.GetServerCount())
	}

	//test removing server
	lb.RemoveServer(newServer.Address)

	if lb.GetServerCount() != 2 {
		t.Errorf("Expected 2 servers after removing one, got %d", lb.GetServerCount())
	}
}

func TestAlgorithmManagement(t *testing.T) {
	servers, testServers := createTestServers(2, true)
	defer cleanup(testServers)

	lb := NewLoadBalancer(servers, "round-robin")

	//test initial algorithm
	if lb.GetAlgorithm() != "round-robin" {
		t.Errorf("Expected round robin, got %s", lb.GetAlgorithm())
	}

	//test changing algorithm
	lb.SetAlgorithm("least-connections")
	if lb.GetAlgorithm() != "least-connections" {
		t.Errorf("Expected to get least-connections, got %s", lb.GetAlgorithm())
	}

	//test invalid algorithm, should remain round-robin
	lb.SetAlgorithm("super-fluous")
	if lb.GetAlgorithm() != "least-connections" {
		t.Errorf("Expected to remain least-connections,got %s", lb.GetAlgorithm())
	}
}

//Test Health Checking

func TestHealthCheck(t *testing.T) {
	//Create servers, one healthy and one unhealthy
	healthyServer := createMockServer("healthy", http.StatusOK, 0)
	unhealthyServer := createMockServer("unhealthy", http.StatusServiceUnavailable, 0)
	defer healthyServer.Close()
	defer unhealthyServer.Close()

	url1, _ := url.Parse(healthyServer.URL)
	url2, _ := url.Parse(unhealthyServer.URL)

	servers := []*Server{
		{Address: healthyServer.URL, IsHealthy: false, URL: url1},  //initially set to false
		{Address: unhealthyServer.URL, IsHealthy: true, URL: url2}, //initially set to true
	}

	//run single health check
	checkAllServers(servers)

	//check results
	if !servers[0].IsHealthy {
		t.Error("Expected first server to be healthy after health check")
	}
	if servers[1].IsHealthy {
		t.Error("Expected second server to be unhealthy after the health check")
	}
}

func TestHealthCheckWithTimeout(t *testing.T) {
	//create slow server that takea longer than health check timeout
	slowServer := createMockServer("slow", http.StatusOK, 10*time.Second)
	defer slowServer.Close()

	slowURL, _ := url.Parse(slowServer.URL)
	servers := []*Server{
		{Address: slowServer.URL, IsHealthy: true, URL: slowURL},
	}

	//run health check (should timeout and mark as unhealthy)
	start := time.Now()
	checkAllServers(servers)
	duration := time.Since(start)

	//Should complete quickly due to timeout
	if duration > 7*time.Second {
		t.Errorf("Health check took too long: %v", duration)
	}

	if servers[0].IsHealthy {
		t.Errorf("Expected slow server to be marked unhealthy due timeout")
	}
}

func TestGetServerCount(t *testing.T) {
	servers, testServers := createTestServers(5, true)
	defer cleanup(testServers)

	//make some servers healthy
	servers[1].IsHealthy = true
	servers[3].IsHealthy = true

	healthy, unhealthy := GetServerCount(servers)

	if healthy != 2 {
		t.Errorf("Expected 2 healthy servers, got %d", healthy)
	}
	if unhealthy != 3 {
		t.Errorf("Expected 3 unhealthy servers, got %d", unhealthy)
	}
}

func TestGetHealthyUnhealthyServers(t *testing.T) {
	servers, testServers := createTestServers(4, false)
	defer cleanup(testServers)

	//make 2 servers healthy
	servers[1].IsHealthy = true
	servers[3].IsHealthy = true

	healthyServers := GetHealthyServers(servers)
	unhealthyServers := GetUnhealthyServers(servers)

	if len(healthyServers) != 2 {
		t.Errorf("Expected to get 2 healthy servers, got %d", len(healthyServers))
	}
	if len(unhealthyServers) != 2 {
		t.Errorf("Expected to get 3 unhealthy servers, got %d", len(unhealthyServers))
	}

	//verify correct servers are returned
	healthyAddresses := []string{healthyServers[1].Address, healthyServers[3].Address}
	expectedhealthyAddresses := []string{servers[1].Address, servers[3].Address}

	for _, expected := range expectedhealthyAddresses {
		found := false
		for _, actual := range healthyAddresses {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected healthy server %s,not found", expected)
		}
	}
}
