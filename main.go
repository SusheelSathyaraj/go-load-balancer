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

func main() {
	fmt.Println("load balancer")
}
