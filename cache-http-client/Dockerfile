# Use the official Golang image from Docker Hub
FROM golang:1.22

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files to the container
COPY go.mod go.sum ./
RUN go mod download

# Copy all .go files to the container
COPY *.go ./

# Build the Go application
RUN go build -o app

# Command to run the executable
CMD ["./app"]
