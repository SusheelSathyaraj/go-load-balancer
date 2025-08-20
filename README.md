# Go Load Balancer

A highly efficient, production-ready load balancer implemented in Go with support for multiple balancing algorithms and comprehensive health checking.

## Features

- **Multiple Load Balancing Algorithm**
  - **Round-robin distibution** 
  - **Least Connections routing**
- **Health Monitoring** 
  - Automatic Health Checks with configurable intervals
  - Real-time server status tracking
  - Graceful handling of Server failures
- **High Performance**
  - Thread-safe concurrent operations
  - Minimal Latency overhead
  - Efficient HTTP proxying
- **Production Ready**
  - Graceful shutdown handling
  - Comprehensive error handling
  - Resource leak prevention
- **Monitoring and Observability**
  - Real-time status endpoint
  - Connection count tracking
  - Health Status reporting
- **Easy Configuration** 
  - YAML based configuration
  - Dynamic server pool management
  - Hot algorithm switching

## Architecture

<img width="1764" height="1084" alt="image" src="https://github.com/user-attachments/assets/557594e2-c778-43d0-a94a-a537fdab3a76" />

## Requirements

- Go 1.21 or higher
- Git
- Compatible with any backend server pool

## Installation

Clone this repository and navigate to the project directory:

```bash
git clone https://github.com/SusheelSathyaraj/go-load-balancer.git
cd go-load-balancer

go mod tidy
```
## Running the Load Balancer
### Method 1: Manual Setup

```bash
# Terminal 1 - Start Backend Server 1
cd server1
go run server1.go

# Terminal 2 - Start Backend Server 2  
cd server2
go run server2.go

# Terminal 3 - Start Load Balancer
go run *.go
```
### Method 2: Using Make

```bash
# Start all servers and load balancer
make start-all

# Or step by step:
make start-servers    # Start backend servers
make run             # Start load balancer
```
## Testing Load Balancer

```bash
# Test load balancing
curl http://localhost:8080/

# Check server status
curl http://localhost:8080/status

# Load test with multiple requests
for i in {1..10}; do curl http://localhost:8080/; echo; done
```
## Configuration

Edit `config.yaml` to customise the setup

```yaml
servers:
  - address: "http://localhost:8081"
  - address: "http://localhost:8082"
  - address: "http://localhost:8083"  # Add more servers
health_check_interval: 10  # Health check interval in seconds
load_balancing_algorithm: "round-robin"  # or "least-connections"
```
## Testing
### Run all tests

```bash
# Basic test run
go test -v

# With race condition detection
go test -v -race

# With test coverage
go test -v -cover

# Generate detailed coverage report
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```
### Run Specific Test Categories

```bash
# Test server functionality
go test -v -run TestServer

# Test load balancing algorithms
go test -v -run TestRoundRobin
go test -v -run TestLeastConnections

# Test health checking
go test -v -run TestHealth

# Test configuration loading
go test -v -run TestLoadConfig

# Test concurrent operations
go test -v -run TestConcurrent

# Integration tests
go test -v -run TestFullIntegration
```
### Performance Benchmark

```bash
# Run all benchmarks
go test -bench=. -v

# Benchmark specific algorithms
go test -bench=BenchmarkRoundRobin -v
go test -bench=BenchmarkLeastConnections -v

# Memory allocation benchmarks
go test -bench=. -benchmem -v

# CPU profiling
go test -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

## Test Coverage Goals
Our Test Suite achieves
  - **95% line coverage**
  - **Concurrent safety validation**
  - **Edge case handling**
  - **Performance Benchmarking**

## API Endpoints



## Load Balancing Algorithms

- Round-robin: Distributes requests sequentially among servers.
- Least Connections: Directs traffic to the server with the fewest active connections.

## Contributing

Contributions are welcome! Please fork this repository and submit a pull request for any changes you’d like to make.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contact

For any queries, please open an issue or reach out via email.
