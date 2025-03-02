.PHONY: build run-rest run-storage run-test docker-build docker-up docker-down clean

# Build all binaries
build:
	go build -o bin/rest-server ./cmd/rest-server
	go build -o bin/storage-server ./cmd/storage-server
	go build -o bin/test-client ./cmd/test-client

# Run the REST server
run-rest:
	go run ./cmd/rest-server/main.go

# Run a storage server
run-storage:
	go run ./cmd/storage-server/main.go --port=8081 --id=storage-1 --rest=http://localhost:8080

# Run the test client
run-test:
	go run ./cmd/test-client/main.go --rest=http://localhost:8080 --storage=http://localhost:8081

# Build Docker images
docker-build:
	docker-compose build

# Start Docker containers
docker-up:
	docker-compose up

# Stop Docker containers
docker-down:
	docker-compose down

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf test-data/ 