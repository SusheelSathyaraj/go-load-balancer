package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	fmt.Println("this is server2")

	port := "8082"

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Server2 is healthy")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//Simulate some procesing time
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Response from Server2 (port %s) at %s", port, time.Now().Format("15:04:05"))
		log.Printf("Handled request on Server2")
	})

	log.Printf("Starting the Server2 on port: %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Server2 failed to start at %s, error: %v", port, err)
	}
}
