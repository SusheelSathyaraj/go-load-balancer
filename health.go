package main

import (
	"net/http"
	"time"
)

func HealthCheck(servers []*Server, interval time.Duration) {
	for {
		for _, server := range servers {
			go func(s *Server) {
				//sending a HTTP GET request to the server’s health endpoint
				resp, err := http.Get(s.Address + "/health")

				s.Mutex.Lock()
				defer s.Mutex.Unlock()

				if err != nil || resp.StatusCode != http.StatusOK {
					s.IsHealthy = false
				}
				s.IsHealthy = true
			}(server)
		}
		time.Sleep(interval)
	}
}
