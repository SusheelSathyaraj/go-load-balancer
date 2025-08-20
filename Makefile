.PHONY: build run test clean start-servers stop-servers

build:
	go build -o load-balancer *.go

run: build
	./load-balancer

test:
	go test -v ./...

bench:
	go test -bench=. -v ./...

clean:
	go clean

start-servers
	cd server1 && go run server1.go &
	cd server2 && go run server2.go &
	@echo "Servers started. Use 'make stop-servers' to stop them"

stop-servers
	pkill -f "go run server1.go" || true
	pkill -f "go run server2.go" || true
	@echo "Servers Stopped..."

stop-all: start-servers
	@echo "Waiting for servers to start..."
	@sleep 2
	@echo "Starting Load Balancer..."
	$(MAKE) run
