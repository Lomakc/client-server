# client-server
Here is a complete, production-ready Go TCP client-server implementation with graceful shutdown, connection handling, and message framing.


# Go TCP Server & Client

A lightweight, concurrent TCP echo server and interactive CLI client built with standard Go packages (`net`, `bufio`, `sync`, `os/signal`).

## Features

- **Concurrent Connections:** Handles multiple clients simultaneously using Go goroutines.
- **Graceful Shutdown:** Server intercepts `SIGINT`/`SIGTERM` to safely drain and complete active connection routines via `sync.WaitGroup`.
- **I/O Deadlines:** Read and write timeouts prevent lingering zombie connections.
- **Framing:** Line-delimited stream reading (`\n`) via `bufio.Reader`.

---

## Getting Started

### 1. Initialize Go Module

```bash
go mod init tcp-server-client
