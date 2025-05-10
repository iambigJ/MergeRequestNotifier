.PHONY: build run dev clean test

# Build the application
build:
	go build -o bin/webhook-server -v

# Run the application
run: build
	./bin/webhook-server

# Run the application with live reloading (requires Air: https://github.com/cosmtrek/air)
dev:
	go install github.com/cosmtrek/air@latest
	air

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Run tests
test:
	go test -v ./...

# Setup project
setup:
	mkdir -p bin
	cp .env.example .env
	go mod tidy 