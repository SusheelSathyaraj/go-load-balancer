package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("load balancer")

	servers := []*Server{
		{Address: "http://localhost:8081", IsHealthy: true},
		{Address: "http://localhost:8082", IsHealthy: true},
	}

	lb := NewLoadBalancer(servers, "round-robin")

	//	Start Health Checks
	go HealthCheck(lb.Servers, 10*time.Second)

	//simulating traffic
	for i := 0; i < 10; i++ {
		server := lb.GetNextServer()
		if server != nil {
			fmt.Printf("Forwarding request to %s\n", server.Address)
		} else {
			fmt.Printf("No healthy servers available!")
		}
		time.Sleep(1 * time.Second)
	}
}
