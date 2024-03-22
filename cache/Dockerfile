FROM golang:1.18

WORKDIR /app

COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

# EXPOSE 5001

RUN go build -o app

CMD ["./app"]