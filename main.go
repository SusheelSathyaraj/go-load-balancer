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
	//waiting for signal
	<-ctx.Done()

	//Graceful shutdown
	log.Println("Shutting down gracefully")
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to shutdown gracefully: %v", err)
	}
	log.Println("Loadbalancer stopped")

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

			//simulate variable load
			go func(s *Server){
				requestTime := time.Duration(100+rand.Intn(400))*time.Millisecond
				time.Sleep(requestTime)
				//simulate request starting
				s.Mutex.Lock()
				s.ConCount++
				s.Mutex.Unlock()

				//simulate request completion
				time.Sleep(500*time.Millisecond)
				s.Mutex.Lock()
				s.ConCount--
				s.Mutex.Unlock()
			}(server)
			time.Sleep(500*time.Millisecond)
		}
	}
			
		} else {
			log.Printf("No healthy servers available!")
		}
		time.Sleep(1 * time.Second)
	}
}
