package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
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

	//starting the loadbalancer on port 8080
	go func() {
		log.Println("Loadbalancer is running on port 8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("failed to start the server %v", err)
		}
	}()

	//simulating traffic in a separate goroutine
	go simulateTraffic(lb)

	//block the main goroutine
	select {}
}

// simulating traffic
func simulateTraffic(lb *Balancer) {
	log.Println("Load Balancer is running. Simulating traffic...")
	for i := 0; i < 20; i++ {
		server := lb.GetNextServer()
		if server != nil {
			log.Printf("Forwarding request %d to %s\n", i+1, server.Address)

			//simulate request starting
			server.Mutex.Lock()
			server.ConCount++
			server.Mutex.Unlock()

			//simulating request completion
			go func(s *Server) {
				time.Sleep(500 * time.Millisecond) // request time
				s.Mutex.Lock()
				s.ConCount--
				s.Mutex.Unlock()
			}(server)
		} else {
			log.Printf("No healthy servers available!")
		}
		time.Sleep(1 * time.Second)
	}
}
