package main

import (
	"fmt"
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
	switch lb.Algo {
	case "round-robin":
		return lb.GetNextServerRoundRobin()
	case "least-connections":
		return lb.GetNextServerLL()
	default:
		fmt.Println("unknown algorithm")
	}

	return nil
}

func (lb *Balancer) GetNextServerRoundRobin() *Server {
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

func (lb *Balancer) GetNextServerLL() *Server {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()

	var selectedServer *Server
	minConnections := int(^uint(0) >> 1)

	for _, server := range lb.Servers {
		server.Mutex.Lock()
		isHealthy := server.IsHealthy
		activeconnections := server.ConCount
		server.Mutex.Unlock()

		if isHealthy && activeconnections < minConnections {
			selectedServer = server
			minConnections = activeconnections
		}
	}
	return selectedServer
}
