FROM golang:1.21-alpine

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Install SQLite dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Build the application
RUN go build -o web_app

# Create template directory
RUN mkdir -p /app/templates

EXPOSE 8080

CMD ["./web_app"]
