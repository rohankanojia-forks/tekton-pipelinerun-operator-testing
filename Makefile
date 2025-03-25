# Variables
APP_NAME := tekton-operator
IMAGE := rohankanojia/$(APP_NAME):latest

.PHONY: all build docker-build docker-push deploy clean

# Default target
all: build

# Build the Go binary
build:
	go mod tidy
	CGO_ENABLED=0 go build -o $(APP_NAME) main.go

# Build the Docker image
docker-build: build
	docker build -t $(IMAGE) .

# Deploy the application to Kubernetes
deploy:
	kubectl apply -f deployment.yaml

# Clean up build files
clean:
	rm -f $(APP_NAME)
