package main

import (
	"sync"
)

// Backend server
type Server struct {
	Address   string
	IsHealthy bool
	Mutex     sync.Mutex
	ConCount  int //for least connection algo
}
