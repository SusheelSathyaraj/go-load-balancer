package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// struct for reading yaml config file
type Config struct {
	Servers                []Server `yaml:"servers"`
	HealthCheckInterval    int      `yaml:"health_check_interval"`
	LoadBalancingAlgorithm string   `yaml:"load_balancing_algorithm"`
}

// struct for the Server type
type Server struct {
	Address   string
	IsHealthy bool
	Concount  int
	Mutex     sync.Mutex
}

func main() {
	//parsing the config file
	config := Config{}
	configData, err := os.ReadFile("config.yaml")
	if err != nil {
		fmt.Printf("Error reading config file: %v", err)
		return
	}
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		fmt.Printf("Error parsing the config file: %v", err)
		return
	}

	//initialising the server and the laodbalancer
	servers := make([]*Server, len(config.Servers))
	for i, srv := range config.Servers {
		servers[i] = &Server{
			Address:   srv.Address,
			IsHealthy: true,
		}
	}

	http.HandleFunc("/simulateLoad", func(w http.ResponseWriter, r *http.Request) {

	})
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Println("server is healthy")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Server is running on port %s", port)
	})

	fmt.Printf("starting server at port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("error starting server on port %s: %v\n", port, err)
	}
}
