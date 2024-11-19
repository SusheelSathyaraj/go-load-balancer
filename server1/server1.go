package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("this is server 1")
	port := "8081"

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
