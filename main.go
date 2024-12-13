package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Address string `yaml:"address"`
}

type Config struct {
	Servers              []ServerConfig `yaml:"servers"`
	HealthCheckIntervals int            `yaml:"health_check_interval"`
	LoadBalancingAlgo    string         `yaml:"load_balancing_algorithm"`
}

// function for loading the config.yaml file
func loadConfig(file string) (*Config, error) {

	configFile, err := os.Open(file)
	if err != nil {
		log.Printf("Error: error reading the config file: %v", err)
		return nil, fmt.Errorf("error reading config file: %v", err)
	}
	defer configFile.Close()

	var config Config
	if err := yaml.NewDecoder(configFile).Decode(&config); err != nil {
		log.Printf("Error: decoding yaml file: %v", err)
		return nil, fmt.Errorf("error decoding yaml file: %v", err)
	}
	return &config, nil
}

// simulating traffic
func simulateTraffic(lb *Balancer) {
	log.Println("Load Balancer is running. Simulating traffic...")
	for i := 0; i < 20; i++ {
		server := lb.GetNextServer()
		if server != nil {
			log.Printf("Forwarding request %d to %s\n", i+1, server.Address)

			//simulate variable load
			go func(s *Server) {
				requestTime := time.Duration(100+rand.Intn(400)) * time.Millisecond
				time.Sleep(requestTime)
				//simulate request starting
				s.Mutex.Lock()
				s.ConCount++
				s.Mutex.Unlock()

				//simulate request completion
				time.Sleep(500 * time.Millisecond)
				s.Mutex.Lock()
				s.ConCount--
				s.Mutex.Unlock()
			}(server)
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// Initial Server validation for health
func validateServers(servers []*Server) {
	for _, server := range servers {
		resp, err := http.Get(server.Address + "/health")
		if err != nil || resp.StatusCode != http.StatusOK {
			server.IsHealthy = false
			log.Printf("Server %s is initially unhealthy", server.Address)
		} else {
			server.IsHealthy = true
			log.Printf("Server %s is initially healthy", server.Address)
		}
	}
}

// simulate traffic to a single server for testing
func simulateTrafficToSingleServer(lb *Balancer, targetAddr string) {
	log.Printf("Simulating traffic for server %s", targetAddr)
	for i := 0; i <= 20; i++ {
		var targetServer *Server

		//Find TargetServer in the loadbalancer
		for _, server := range lb.Servers {
			if server.Address == targetAddr {
				targetServer = server
				break
			}
		}
		if targetServer == nil {
			log.Printf("Target Server %s not present in the loadbalancer", targetAddr)
			return
		}
		log.Printf("Forwarding the request %d to server %s", i+1, targetAddr)

		//simulate starting the request
		targetServer.Mutex.Lock()
		targetServer.ConCount++
		targetServer.Mutex.Unlock()

		//simulate request completion
		go func(s *Server) {
			time.Sleep(500 * time.Millisecond)
			s.Mutex.Lock()
			s.ConCount--
			s.Mutex.Unlock()
		}(targetServer)
		time.Sleep(1 * time.Second)
	}
}

func main() {
	log.Println("load balancer starting")

	//loading the config file
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("failed to load the config file: %v", err)
	}

	servers := make([]*Server, len(config.Servers))
	for i, srv := range config.Servers {
		servers[i] = &Server{Address: srv.Address, IsHealthy: true}
	}

	//loads the loadsbalancer
	lb := NewLoadBalancer(servers, config.LoadBalancingAlgo)

	//Perform validation of the servers
	validateServers(lb.Servers)

	//	Start Health Checks
	go HealthCheck(lb.Servers, time.Duration(config.HealthCheckIntervals)*time.Second)

	//HTTP for loadbalancer
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		server := lb.GetNextServer()
		if server == nil {
			http.Error(w, "no healthy servers present", http.StatusServiceUnavailable)
			return
		}
		//backend server
		resp, err := http.Get(server.Address + r.URL.Path)
		if err != nil {
			http.Error(w, "failed to forward the request", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		//copy the response from the backend to client
		w.WriteHeader(resp.StatusCode)
		resp.Write(w)

	})

	//context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	//start http server
	server := &http.Server{Addr: ":8080", Handler: nil}

	//starting the loadbalancer on port 8080
	go func() {
		log.Println("Loadbalancer is running on port 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start the server %v", err)
		}
	}()

	//waiting briefly to ensure the server is running successfully before simulating traffic
	time.Sleep(5 * time.Second)

	//simulating traffic in a separate goroutine
	//go simulateTraffic(lb)

	//simulate traffic to a single server
	//used here only for testing
	//when used, comment simulateTraffic(lb)

	go simulateTrafficToSingleServer(lb, "http://localhost:8081")

	//waiting for signal
	<-ctx.Done()

	//Graceful shutdown
	log.Println("Shutting down gracefully")
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to shutdown gracefully: %v", err)
	}
	log.Println("Loadbalancer stopped")
}
