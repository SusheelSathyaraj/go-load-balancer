package main

import (
	"time"
)

func HealthCheck(servers []*Server, interval time.Duration) {
	for {
		for _, server := range servers {
			go func(s *Server) {

			}(server)

		}

		time.Sleep(interval)
	}

}
