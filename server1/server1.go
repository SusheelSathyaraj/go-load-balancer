package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	fmt.Println("Starting server1...")

	port := "8081"

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Server1 is healthy")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//Simulate some processing time
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Response from Server1 (port %s) at %s", port, time.Now().Format("15:04:05"))
		log.Printf("Handled request on Server1")
	})

	log.Printf("Starting Server 1 on port: %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server 1 failed to start on port: %s,%v", port, err)
	}
}
