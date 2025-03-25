# Use a minimal base image for Go
FROM golang:1.23 AS builder

# Set working directory
WORKDIR /app

# Copy and build the application
COPY . .
RUN go mod tidy && GOOS=linux GOARCH=amd64 go build -o tekton-operator main.go

# Use a minimal runtime image
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/tekton-operator .

# Run the application
CMD ["./tekton-operator"]
