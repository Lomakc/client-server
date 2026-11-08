package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	address      = ":8080"
	readTimeout  = 60 * time.Second
	writeTimeout = 10 * time.Second
)

type Server struct {
	listener net.Listener
	wg       sync.WaitGroup
	quit     chan struct{}
}

func NewServer(addr string) (*Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Server{
		listener: ln,
		quit:     make(chan struct{}),
	}, nil
}

func (s *Server) Start() {
	log.Printf("Server listening on %s\n", s.listener.Addr().String())

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				log.Printf("Accept error: %v\n", err)
				continue
			}
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	log.Printf("[%s] Client connected\n", remoteAddr)

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		// Set deadline to prevent idle connections hanging forever
		_ = conn.SetReadDeadline(time.Now().Add(readTimeout))

		msg, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Printf("[%s] Client disconnected\n", remoteAddr)
			} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("[%s] Read timeout reached\n", remoteAddr)
			} else {
				log.Printf("[%s] Read error: %v\n", remoteAddr, err)
			}
			return
		}

		cleanMsg := strings.TrimSpace(msg)
		log.Printf("[%s] Received: %s\n", remoteAddr, cleanMsg)

		if strings.EqualFold(cleanMsg, "QUIT") {
			_ = conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			_, _ = writer.WriteString("BYE\n")
			_ = writer.Flush()
			return
		}

		// Echo back with server acknowledgement
		reply := fmt.Sprintf("ACK: %s\n", cleanMsg)
		_ = conn.SetWriteDeadline(time.Now().Add(writeTimeout))
		if _, err := writer.WriteString(reply); err != nil {
			log.Printf("[%s] Write error: %v\n", remoteAddr, err)
			return
		}
		if err := writer.Flush(); err != nil {
			log.Printf("[%s] Flush error: %v\n", remoteAddr, err)
			return
		}
	}
}

func (s *Server) Stop() {
	close(s.quit)
	_ = s.listener.Close()
	s.wg.Wait()
	log.Println("Server stopped cleanly.")
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv, err := NewServer(address)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	go srv.Start()

	<-ctx.Done()
	log.Println("Shutting down server...")
	srv.Stop()
}
