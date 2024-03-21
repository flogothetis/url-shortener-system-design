# Dockerfile

FROM golang:1.17

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY *.go ./

# EXPOSE 5001

RUN go build -o app

CMD ["./app"]