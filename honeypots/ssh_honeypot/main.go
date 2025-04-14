package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// AttackData represents the information captured from an SSH attack attempt
type AttackData struct {
	Timestamp  time.Time `json:"timestamp"`
	SourceIP   string    `json:"source_ip"`
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	ClientVersion string  `json:"client_version"`
}

var (
	logFile *os.File
	mutex   sync.Mutex
)

func main() {
	// Configure logging
	logPath := getEnv("LOG_PATH", "/var/log/honeypot/ssh_attempts.log")
	logDir := getEnv("LOG_DIR", "/var/log/honeypot")

	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	// Open log file
	var err error
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// Configure SSH server
	config := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			// Always fail authentication but log the attempt
			username := conn.User()
			remoteAddr := conn.RemoteAddr().String()
		
			// Extract IP without port
			host, _, _ := net.SplitHostPort(remoteAddr)
		
			// Record the attempt
			logAttempt(AttackData{
				Timestamp:     time.Now(),
				SourceIP:      host,
				Username:      username,
				Password:      string(password),
				ClientVersion: string(conn.ClientVersion()),
			})
		
			return nil, fmt.Errorf("authentication failed")
		},
		ServerVersion: "SSH-2.0-OpenSSH_7.4", // Impersonate a slightly outdated version
	}

	// Generate server key
	privateKey, err := generateHostKey()
	if err != nil {
		log.Fatalf("Failed to generate host key: %v", err)
	}

	// Add private key to server config
	privateKeySigner, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}
	config.AddHostKey(privateKeySigner)

	// Start SSH server
	port := getEnv("SSH_PORT", "2222")
	listener, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}
	defer listener.Close()

	log.Printf("SSH Honeypot listening on port %s...", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		// Handle connections in a goroutine
		go handleConnection(conn, config)
	}
}

// handleConnection processes an incoming SSH connection
func handleConnection(conn net.Conn, config *ssh.ServerConfig) {
	defer conn.Close()

	// Perform SSH handshake
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		log.Printf("Failed to handshake: %v", err)
		return
	}
	defer sshConn.Close()

	log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())

	// Service the incoming channels and requests
	go ssh.DiscardRequests(reqs)
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("Failed to accept channel: %v", err)
			continue
		}

		go func(ch ssh.Channel, reqs <-chan *ssh.Request) {
			defer ch.Close()
			
			for req := range reqs {
				switch req.Type {
				case "shell":
					// Accept request but don't actually create a shell
					req.Reply(true, nil)
					io.WriteString(ch, "Welcome to the server.\nThis incident has been reported.\n")
					ch.Close()
				case "exec":
					// Log the command they're trying to execute
					cmd := string(req.Payload[4:]) // Skip 4-byte length prefix
					log.Printf("Attempted command execution: %s from %s", cmd, sshConn.RemoteAddr())
					req.Reply(true, nil)
					io.WriteString(ch, "Permission denied\n")
					ch.Close()
				default:
					log.Printf("Unknown request type: %s", req.Type)
					req.Reply(false, nil)
				}
			}
		}(channel, requests)
	}
}

// logAttempt records the SSH login attempt to both the log file and stdout
func logAttempt(data AttackData) {
	mutex.Lock()
	defer mutex.Unlock()

	// Marshal the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling attack data: %v", err)
		return
	}

	// Write to log file
	if _, err := logFile.Write(append(jsonData, '\n')); err != nil {
		log.Printf("Error writing to log file: %v", err)
		return
	}

	// Also output to stdout for monitoring
	log.Printf("SSH login attempt: %s from %s with password '%s'", 
		 data.Username, data.SourceIP, data.Password)
}

// generateHostKey creates an in-memory SSH key for the server
func generateHostKey() ([]byte, error) {
	key := `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAvXd/TqKtUUgv4ClUzxOQ2XC8TbFRYyZEBFO6LZ4tdXQxDnXy
KXGQeE3o9BBdS0vGU8leIIfXUzvcMkSqFsTTXGYgW57ZZhpVCBplRMK3zBiBFt4+
RpxyGzF0Bz2jSmtZ2x6pX6+jVHuLQwT2d0bh5qlND/DwZ6aq9NJVXmvgKFdohxnk
qXKRrSN/pppJP9nYrUcL4jHJQkwvGgYbvd8Le3bRs+jazh+vV9KxYR4wQd6U/Wgn
WITc4glkRziQZUCVJnEMwB9vSs+ES5QOViIGNQF4QkgJU802GbwE41D3iBLsQQQv
W+NKX9amlB/BSn0JbFakijj6jUkpG4UXd+QYQQIDAQABAoIBAQC8VIlpm3xfx5Ak
uVYgZq1/6C8p9QFcBeHx2GRXBI7plU7gz9+/vLWU558J5xVKdpSW9WyK67K18woS
m1pBsOTuUCGGpJ4JIt0U/0aJ/f6G2sj1eB/7g59jQx/X8C5kv4YkvUCDQwI/bI2p
dYEcF3nYAKrPMdBEK5Bv4v0j4l6oVCqwGR5OyHgvk0vUNkrdvq7Tba+O4qtY3BHU
VFdYP1Kos27gyFf++eEQUfB/3pYIg7UObpwVlHHOYXJmKC7cK4pREoe+XocP/GYT
0k4CkZjrk9CvK9jCvRyMQmiJQ8bJp5TKiQPvDtiF4xKthGYQqTmLYF3YJRXxCtj/
ZzRnpWDRAoGBAPLMHgkaKD0/7bM9nVzcRuFBVHcPcZQCfCndEn8eyD4tLYmC6gOV
zH0ChaaMZq9NwUBQ4+9jIE4+QhCiY3JzfZkE3/EzK3Q5eTKM5wkFQQQd+beGY3vS
G4+9+8QAodmCxcA4mMkIsUOIP4mtqJ1nO7Ifg1aZKTlU+P14YUOWqXGVAoGBAMgG
CSFkVhhgJCJLkjGZ758WSz6bG9Xx6L60PUw0Qs0TQGHWDKlbWoZOzDdM7U4/+w0X
mTjgYGCqxBbKsl/x/mVbYQXimAKzLWMJyBmMl8Ferx9UavU0GkKrSDtwBYTQkMEp
OUeiSYZZlQ6VSgndZCSTtkWCnUyDQvjJM0ymGdudAoGBAOtLVD6mOj/Z+7uyeRjZ
xGXNNH5ITLI9Y03OgB8g2b3gUzK6A2QbIGIIM4znNt1I20ail9Zj38gy6uLYWNIi
bLgYUtDiG2zPGXyZdZcGqfx/cQTwYk9t6bdN2xYTZquZGtILOAZNFoS7O2ieHlw3
1q22GbYRVFNHy3ZGy8oE6RB1AoGAJnTvTMKcvbPiUAVIiZUZDC8JBjIwWisvkXDK
3svKBnJWciE+CgXEhQ2q86SHjGIwdyEy1McgD1hDuOi5vj0+8/9gnTl4Z1dA1/c+
+kNgOzJRfYZc1PWJ/JJ9QhbEnKSfaUfJm0d/+lyLHuSvjPUb56B6GR+q2MRRuQ77
5VLa990CgYA4QdZ5gFQxVIKDcYs4ks5MiR/FDFN8P31Sol6J14KzCGh6WFnXfw8I
ipzZxDbwM3vQYQ7+umdnznwJqJ6AMVYEpPDrwqpxSQKjqDbOYXUcPgDZLXVy3SaX
wLPsZuNJlJl7JC2lPmK2cZ9WwOCcGDdRMqg2cW3eUyDAWHKtx8IPwQ==
-----END RSA PRIVATE KEY-----`
	return []byte(key), nil
}

// getEnv returns the value of an environment variable or a default value if not set
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
