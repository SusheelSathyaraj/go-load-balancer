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

- Go 1.16 or higher
- Compatible with any backend server pool

## Installation

Clone this repository and navigate to the project directory:

```bash
git clone https://github.com/SusheelSathyaraj/go-load-balancer.git
cd go-load-balancer

go mod tidy
```

## Usage

- Configure Backend Servers: Update config.yaml to specify the servers for load balancing.
- Run the Load Balancer: Execute the following command to start the load balancer:

```bash
  go run *.go -> to run all the .go extension files at once
```
- Access Logs and Monitoring: Check logs in the logs/ folder for detailed information on traffic distribution and health checks.

## Configuration

Modify the config.yaml file to specify server details and health check intervals. Example:

```yaml
servers:
  - address: http://server1.com
  - address: http://server2.com
healthCheckInterval: 10s
loadBalancingAlgorithm: round-robin and least connections
```

## Load Balancing Algorithms

- Round-robin: Distributes requests sequentially among servers.
- Least Connections: Directs traffic to the server with the fewest active connections.

## Contributing

Contributions are welcome! Please fork this repository and submit a pull request for any changes you’d like to make.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contact

For any queries, please open an issue or reach out via email.
