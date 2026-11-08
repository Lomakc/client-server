package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const (
	serverAddr   = "127.0.0.1:8080"
	dialTimeout  = 5 * time.Second
	writeTimeout = 5 * time.Second
)

func main() {
	log.Printf("Connecting to %s...\n", serverAddr)
	conn, err := net.DialTimeout("tcp", serverAddr, dialTimeout)
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer conn.Close()
	log.Println("Connected. Type messages to send, or 'QUIT' to exit.")

	stdinScanner := bufio.NewScanner(os.Stdin)
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		fmt.Print("> ")
		if !stdinScanner.Scan() {
			break
		}

		text := strings.TrimSpace(stdinScanner.Text())
		if text == "" {
			continue
		}

		// Send message to server
		_ = conn.SetWriteDeadline(time.Now().Add(writeTimeout))
		if _, err := writer.WriteString(text + "\n"); err != nil {
			log.Fatalf("Failed to write to server: %v", err)
		}
		if err := writer.Flush(); err != nil {
			log.Fatalf("Failed to flush to server: %v", err)
		}

		// Read response
		response, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("Server closed connection.")
				return
			}
			log.Fatalf("Failed to read server response: %v", err)
		}

		cleanResp := strings.TrimSpace(response)
		fmt.Printf("< %s\n", cleanResp)

		if strings.EqualFold(text, "QUIT") || cleanResp == "BYE" {
			log.Println("Session terminated.")
			return
		}
	}

	if err := stdinScanner.Err(); err != nil {
		log.Printf("Input scanner error: %v", err)
	}
}
