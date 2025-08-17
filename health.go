package main

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

// performing periodic health checks
func HealthCheck(servers []*Server, interval time.Duration, ctx context.Context) {
	log.Printf("Starting health checks with %v interval", interval)

	//Initial Health Checks
	checkAllServers(servers)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping health checks")
			return
		case <-ticker.C:
			checkAllServers(servers)
		}
	}
}

// performing health check on all the servers concurrently
func checkAllServers(servers []*Server) {
	var wg sync.WaitGroup

	for _, server := range servers {
		wg.Add(1)
		go func(s *Server) {
			defer wg.Done()
			checkserverHealth(s)
		}(server)
	}
	wg.Wait()
}

// performing health check on a single server
func checkserverHealth(server *Server) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	//send http get request to the server's health endpoint
	resp, err := client.Get(server.Address + "/health")

	server.Mutex.Lock()
	defer server.Mutex.Unlock()

	previousHealth := server.IsHealthy

	if err != nil || resp.StatusCode != http.StatusOK {
		server.IsHealthy = false
		if previousHealth {
			log.Printf("Server %s is unhealthy ,%v", server.Address, err)
		}
	} else {
		server.IsHealthy = true
		if !previousHealth {
			log.Printf("Server %s is healthy", server.Address)
		}
		resp.Body.Close()
	}
}

// performing a one-time health check on all servers
func PerformSingleHealthCheck(servers []*Server) {
	log.Println("Performing initial health check...")
	checkAllServers(servers)

	healthycount := 0

	for _, server := range servers {
		server.Mutex.Lock()
		if server.IsHealthy {
			healthycount++
		}
		server.Mutex.Unlock()
	}
	log.Printf("Initial health check completed: %d of %d servers healthy", healthycount, len(servers))
}

// Returning the helath status of all the servers
func GetHealthStatus(servers []*Server) map[string]bool {
	status := make(map[string]bool)

	for _, server := range servers {
		server.Mutex.Lock()
		status[server.Address] = server.IsHealthy
		server.Mutex.Unlock()
	}
	return status
}

// Checking if atleast one server is healthy
func IsAnyServerHealthy(servers []*Server) bool {
	for _, server := range servers {
		server.Mutex.Lock()
		isHealthy := server.IsHealthy
		server.Mutex.Unlock()

		if isHealthy {
			return true
		}
	}
	return false
}

// getting a list of healthy servers
func GetHealthyServers(servers []*Server) []*Server {
	var healthy []*Server

	for _, server := range servers {
		server.Mutex.Lock()
		if server.IsHealthy {
			healthy = append(healthy, server)
		}
		server.Mutex.Unlock()
	}
	return healthy
}

// getting the number of healthy and unhealthy servers in a single pass
func GetServerCount(servers []*Server) (healhty, unhealthy int) {
	healthy, unhealthy := 0, 0
	for _, server := range servers {
		server.Mutex.Lock()
		if server.IsHealthy {
			healthy++
		} else {
			unhealthy++
		}
		server.Mutex.Unlock()
	}
	return healthy, unhealthy
}

// getting a list of unhealthy servers
func GetUnhealthyServers(servers []*Server) []*Server {
	var unhealthy []*Server

	for _, server := range servers {
		server.Mutex.Lock()
		if !server.IsHealthy {
			unhealthy = append(unhealthy, server)
		}
		server.Mutex.Unlock()
	}
	return unhealthy
}
