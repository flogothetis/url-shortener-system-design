# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.21 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

# Build the application binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-gs-ping

# Run tests in a separate stage
FROM build-stage AS run-test-stage


# Deploy the application binary into a lean image
FROM alpine:latest AS build-release-stage

WORKDIR /

# Copy the built application binary from build-stage
COPY --from=build-stage /docker-gs-ping /docker-gs-ping

# Expose the port
EXPOSE 3001

# Set the user
USER nobody:nobody

# Entrypoint for running the application
ENTRYPOINT ["/docker-gs-ping"]
