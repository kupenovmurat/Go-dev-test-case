.PHONY: build run-rest run-storage run-test docker-build docker-up docker-down clean test install-deps help

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
	rm -rf data/
	rm -rf logs/

# Run tests
test:
	go test ./...

# Install dependencies
install-deps:
	go mod download

# Show help
help:
	@echo "Доступные команды:"
	@echo "  make build         - Сборка всех бинарных файлов"
	@echo "  make run-rest      - Запуск REST-сервера"
	@echo "  make run-storage   - Запуск сервера хранения"
	@echo "  make run-test      - Запуск тестового клиента"
	@echo "  make docker-build  - Сборка Docker-образов"
	@echo "  make docker-up     - Запуск через Docker Compose"
	@echo "  make docker-down   - Остановка Docker Compose"
	@echo "  make clean         - Удаление временных файлов"
	@echo "  make test          - Запуск тестов"
	@echo "  make install-deps  - Установка зависимостей"
	@echo "  make help          - Показать эту справку" 