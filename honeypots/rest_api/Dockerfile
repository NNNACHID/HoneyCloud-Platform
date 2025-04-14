FROM golang:1.21-alpine

WORKDIR /app

COPY . .

RUN go mod init honeypot/rest_api && \
    go mod tidy && \
    go build -o rest_api

EXPOSE 8080

CMD ["./rest_api"]
