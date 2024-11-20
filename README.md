# Go Load Balancer

A highly efficient, custom-built load balancer implemented in Go, designed to evenly distribute incoming network traffic across multiple servers. This load balancer helps ensure high availability and optimal resource utilization, making it ideal for scalable and fault-tolerant applications.

## Features

- **Round-robin** and **Least Connections** balancing algorithms
- **Health Checks** to ensure only available servers receive traffic
- **Dynamic Server Pool** for easy addition and removal of servers
- **Error Logging** and **Monitoring** capabilities
- **High Performance** with minimal latency

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
  go run main.go
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
