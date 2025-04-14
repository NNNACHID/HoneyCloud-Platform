FROM golang:1.16

WORKDIR /app

COPY . .

RUN go build -o ssh_honeypot

CMD ["./ssh_honeypot"]
