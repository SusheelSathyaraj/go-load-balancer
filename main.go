package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	fmt.Println("load balancer")

	servers := []*Server{
		{Address: "http://localhost:8081", IsHealthy: true},
		{Address: "http://localhost:8082", IsHealthy: true},
	}

	algo := "least-connections" //change to round robin or other algos as required
	lb := NewLoadBalancer(servers, algo)

	//	Start Health Checks
	go HealthCheck(lb.Servers, 10*time.Second)

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

		//starting the loadbalancer on port 8080
		fmt.Println("Loadbalancer is running on port 8080")
		http.ListenAndServe(":8080", nil)
	})

	//simulating traffic
	fmt.Println("Load Balancer is running. Simulating traffic...")
	for i := 0; i < 20; i++ {
		server := lb.GetNextServer()
		if server != nil {
			fmt.Printf("Forwarding request %d to %s\n", i+1, server.Address)

			//simulating starting a server
			go func(s *Server) {
				time.Sleep(500 * time.Millisecond) // request time
				s.Mutex.Lock()
				s.ConCount++
				s.Mutex.Unlock()
			}(server)
		} else {
			fmt.Printf("No healthy servers available!")
		}
		time.Sleep(1 * time.Second)
	}
}
