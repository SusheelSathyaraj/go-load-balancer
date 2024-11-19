package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("this is server2")

	port := "8082"

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Server is healthy")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Server is running on port: %s", port)
	})

	fmt.Printf("Starting the server on port: %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Server is not running on port %s, error: %v", port, err)
	}
}
