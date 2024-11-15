package main

import (
	"sync"
)

type Balancer struct {
	Servers []*Server
	Current int
	Mutex   sync.Mutex
	Algo    string //selection between round robin and least connection
}

// loadbalancer code
func NewLoadBalancer(server []*Server, algo string) *Balancer {
	return &Balancer{
		Servers: server,
		Current: 0,
		Algo:    algo,
	}
}

func (lb *Balancer) GetNextServer() *Server {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()

	for i := 0; i < len(lb.Servers); i++ {
		idx := lb.Current % len(lb.Servers)
		lb.Current++

		server := lb.Servers[idx]

		server.Mutex.Lock()
		isHealthy := server.IsHealthy
		server.Mutex.Unlock()

		if isHealthy {
			return server
		}
	}
	return nil
}
