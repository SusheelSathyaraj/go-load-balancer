package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
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

	//setting defaults
	if config.HealthCheckIntervals == 0 {
		config.HealthCheckIntervals = 10
	}
	if config.LoadBalancingAlgo == "" {
		config.LoadBalancingAlgo = "round-robin"
	}

	return &config, nil
}

// HTTP handler for load balancing
func (lb *Balancer) handleRequest(w http.ResponseWriter, r *http.Request) {
	server := lb.GetNextServer()
	if server == nil {
		http.Error(w, "No healthy servers available", http.StatusServiceUnavailable)
		return
	}

	//increment connection count
	server.Mutex.Lock()
	server.ConCount++
	server.Mutex.Unlock()

	//decrement connection count when completed
	defer func() {
		server.Mutex.Lock()
		server.ConCount--
		server.Mutex.Unlock()
	}()

	//creating a proxy request
	proxyReq, err := http.NewRequest(r.Method, server.Address+r.URL.Path, r.Body)
	if err != nil {
		http.Error(w, "failed to create proxy request", http.StatusInternalServerError)
		return
	}

	//copying headers
	for header, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(header, value)
		}
	}

	//making request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, "failed to forward request", http.StatusBadGateway)
		log.Printf("Failed to forward request to %s: %v", server.Address, err)
		return
	}
	defer resp.Body.Close()

	//copying respose headers
	for header, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	//copying status code
	w.WriteHeader(resp.StatusCode)

	//copying resposnse body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("error copying the response body, %v", err)
	}

	log.Printf("Request forwarded to %s, status: %d", server.Address, resp.StatusCode)
}

// handler for status endpoint
func (lb *Balancer) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"healthy","algorithm":"%s","servers":[`, lb.Algo)

	for i, server := range lb.Servers {
		server.Mutex.Lock()
		if i > 0 {
			fmt.Fprintf(w, ",")
		}
		fmt.Fprintf(w, `{"address":"%s","healthy":"%v","connections":"%d"}`, server.Address, server.IsHealthy, server.ConCount)
		server.Mutex.Unlock()
	}
	fmt.Fprintf(w, `]}`)
}

// simulating traffic for testing
func simulateTraffic(ctx context.Context) {
	log.Println("Load Balancer is running. Simulating traffic...")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for i := 0; i < 50; i++ {
		select {
		case <-ctx.Done():
			log.Println("Stopping traffic simulation")
			return
		default:
			go func(requestID int) {
				resp, err := client.Get("http://localhost:8080/")
				if err != nil {
					log.Printf("Request %d failed %v", requestID, err)
					return
				}
				defer resp.Body.Close()

				body, _ := io.ReadAll(resp.Body)
				bodyStr := string(body)
				if len(bodyStr) > 50 {
					bodyStr = bodyStr[:50] + "..."
				}
				log.Printf("Request %d completed: %s", requestID, bodyStr)
			}(i + 1)

			//Random delay between requests
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		}
	}
}

func main() {
	log.Println("Load balancer starting...")

	//loading the config file
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("failed to load the config file: %v", err)
	}

	//initializing servers
	servers := make([]*Server, len(config.Servers))
	for i, srv := range config.Servers {
		serverURL, err := url.Parse(srv.Address)
		if err != nil {
			log.Fatalf("Invalid server URL %s: %v", srv.Address, err)
		}
		servers[i] = &Server{Address: srv.Address, IsHealthy: false, URL: serverURL}
	}

	//creates the loadsbalancer using loadbalancer.go
	lb := NewLoadBalancer(servers, config.LoadBalancingAlgo)

	log.Printf("Load Balancer configured with %d servers using %s algorithm", len(servers), config.LoadBalancingAlgo)

	//context for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	//	Start Health Checks
	go HealthCheck(lb.Servers, time.Duration(config.HealthCheckIntervals)*time.Second, ctx)

	//wait for initial healthchecks
	time.Sleep(2 * time.Second)

	//setting up HTTP server with method binding
	http.HandleFunc("/", lb.handleRequest)
	http.HandleFunc("/status", lb.handleStatus)

	//starting HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: nil,
	}

	//starting the loadbalancer on port 8080
	go func() {
		log.Println("Loadbalancer is running on port 8080")
		log.Println("Endpoints: http://localhost:8080/ (load balanced), http://localhost:8080/status (status)")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start the server %v", err)
		}
	}()

	//waiting briefly to ensure the server is running successfully before simulating traffic
	time.Sleep(5 * time.Second)

	//simulating traffic in a separate goroutine
	go simulateTraffic(ctx)

	//waiting for signal
	<-ctx.Done()

	//Graceful shutdown
	log.Println("Shutting down loadbalancer")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Failed to shutdown gracefully: %v", err)
	}
	log.Println("Loadbalancer stopped successfully")
}
