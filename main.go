package main

import (
	"fmt"
	"net/url"
	"sync"
)

type loadBalancer struct {
	Current int
	Mutex   sync.Mutex
}

type Server struct {
	URL       *url.URL
	IsHealthy bool
	Mutex     sync.Mutex
}

// round robin logic
func (lb *loadBalancer) getNextServer(servers []*Server) *Server {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()

	for i := 0; i < len(servers); i++ {
		idx := lb.Current % len(servers)
		nextServer := servers[idx]
		lb.Current++

		nextServer.Mutex.Lock()
		isHealthy := nextServer.IsHealthy
		nextServer.Mutex.Unlock()

		if isHealthy {
			return nextServer
		}
	}
	return nil
}

func main() {
	fmt.Println("load balancer")
}
