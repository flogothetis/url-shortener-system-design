# Dockerfile

FROM golang:1.17

WORKDIR /app

COPY . .

RUN go mod init gosnowflake
RUN go mod tidy

RUN go build -o app

CMD ["./app"]