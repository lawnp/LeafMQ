# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get

# Main app
MAIN = cmd/main.go
BINARY_NAME = bin/mqtt-broker

# Packages
PACKAGES = $(wildcard pkg/*)

# Build the main app
build:
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN)

# Clean the project
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Install dependencies
deps:
	$(GOGET) -v ./...

# Run tests
test:
	$(GOTEST) -v ./...

# Build and run the main app
run: build
	./$(BINARY_NAME)

# Build the main app and install dependencies
build-all: deps build

# Build and run the main app with dependencies
run-all: build-all
	./$(BINARY_NAME)

# Default target
default: build
