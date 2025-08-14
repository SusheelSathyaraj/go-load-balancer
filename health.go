package main

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

func HealthCheck(servers []*Server, interval time.Duration, ctx context.Context) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	//Initial Health Checks
	checkAllServers(servers)

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

func checkserverHealth(server *Server) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(server.Address + "/health")

	server.Mutex.RLock()
	defer server.Mutex.RUnlock()

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
