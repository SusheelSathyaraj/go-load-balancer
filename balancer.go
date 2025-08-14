package main

import (
	"log"
	"sync"
)

type Balancer struct {
	Servers []*Server
	Current int
	Mutex   sync.RWMutex
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
	switch lb.Algo {
	case "round-robin":
		return lb.GetNextServerRoundRobin()
	case "least-connections":
		return lb.GetNextServerLL()
	default:
		log.Printf("Unknown algorithm: %s, using round robin", lb.Algo)
		return lb.GetNextServerRoundRobin()
	}
}

func (lb *Balancer) GetNextServerRoundRobin() *Server {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()

	if len(lb.Servers) == 0 {
		return nil
	}

	attempts := 0
	for attempts < len(lb.Servers) {
		idx := lb.Current % len(lb.Servers)
		lb.Current++

		server := lb.Servers[idx]

		server.Mutex.RLock()
		isHealthy := server.IsHealthy
		server.Mutex.RUnlock()

		if isHealthy {
			return server
		}
		attempts++
	}
	return nil
}

func (lb *Balancer) GetNextServerLL() *Server {
	lb.Mutex.RLock()
	defer lb.Mutex.RUnlock()

	var selectedServer *Server
	minConnections := int(^uint(0) >> 1) //max int

	for _, server := range lb.Servers {
		server.Mutex.RLock()
		isHealthy := server.IsHealthy
		activeconnections := server.ConCount
		server.Mutex.RUnlock()

		if isHealthy && activeconnections < minConnections {
			selectedServer = server
			minConnections = activeconnections
		}
	}
	return selectedServer
}
