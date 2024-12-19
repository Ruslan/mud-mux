package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type MUX struct {
	mudAddr   string
	localAddr string
	clients   []net.Conn
	mu        sync.Mutex
	mudConn   net.Conn
	logger    *log.Logger
}

func NewMUX(mudAddr, localAddr, logFile string) *MUX {
	var logger *log.Logger
	if logFile != "" {
		lumberJackLogger := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    500, // megabytes
			MaxBackups: 3,
			MaxAge:     28,   // days
			Compress:   true, // disabled by default
		}
		logger = log.New(io.MultiWriter(os.Stdout, lumberJackLogger), "MUX: ", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		logger = log.New(os.Stdout, "MUX: ", log.Ldate|log.Ltime|log.Lshortfile)
	}

	return &MUX{
		mudAddr:   mudAddr,
		localAddr: localAddr,
		clients:   make([]net.Conn, 0),
		logger:    logger,
	}
}

func (m *MUX) connectToMUD() error {
	conn, err := net.Dial("tcp", m.mudAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to MUD server: %w", err)
	}
	m.mudConn = conn
	m.logger.Printf("Connected to MUD: %s", m.mudAddr)
	return nil
}

func (m *MUX) readFromMUD() {
	buf := make([]byte, 4096)
	for {
		n, err := m.mudConn.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			m.logger.Printf("Error reading from MUD: %v", err)
			return
		}
		m.logger.Printf("MUD -> Clients: %s", buf[:n])
		m.broadcastToClients(buf[:n])
	}
}

func (m *MUX) handleClient(clientConn net.Conn) {
	m.mu.Lock()
	m.clients = append(m.clients, clientConn)
	m.mu.Unlock()

	defer func() {
		clientConn.Close()
		m.mu.Lock()
		for i, c := range m.clients {
			if c == clientConn {
				m.clients = append(m.clients[:i], m.clients[i+1:]...)
				break
			}
		}
		m.mu.Unlock()
		m.logger.Printf("Client disconnected: %s", clientConn.RemoteAddr())
	}()

	m.logger.Printf("New client connected: %s", clientConn.RemoteAddr())

	buf := make([]byte, 4096)
	for {
		n, err := clientConn.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			m.logger.Printf("Error reading from client: %v", err)
			return
		}
		m.logger.Printf("Client -> MUD: %s", buf[:n])
		_, err = m.mudConn.Write(buf[:n])
		if err != nil {
			m.logger.Printf("Error sending to MUD: %v", err)
			return
		}
	}
}

func (m *MUX) broadcastToClients(data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, client := range m.clients {
		_, err := client.Write(data)
		if err != nil {
			m.logger.Printf("Error broadcasting to client: %v", err)
		}
	}
}

func (m *MUX) start() error {
	if err := m.connectToMUD(); err != nil {
		return err
	}
	defer m.mudConn.Close()

	go m.readFromMUD()

	listener, err := net.Listen("tcp", m.localAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer listener.Close()

	m.logger.Printf("MUX listening on %s", m.localAddr)

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Handle graceful shutdown
	go func() {
		<-sigchan
		m.logger.Println("Shutting down MUX...")
		m.disconnectAllClients()
		if err := m.mudConn.Close(); err != nil {
			m.logger.Printf("Error closing MUD connection: %v", err)
		}
		os.Exit(0)
	}()

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			m.logger.Println("Accept error:", err)
			time.Sleep(time.Second) // Small delay to prevent excessive logging
			continue
		}
		go m.handleClient(clientConn)
	}
}

func (m *MUX) disconnectAllClients() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, client := range m.clients {
		client.Close()
	}
	m.clients = nil
}

func main() {
	mudAddr := flag.String("mud", "muds.example.com:23", "MUD server address (host:port)")
	localAddr := flag.String("local", ":8888", "Local listening address (host:port)")
	logFile := flag.String("log", "log/game.log", "Log file path (optional, enables logging with rotation)")
	flag.Parse()

	mux := NewMUX(*mudAddr, *localAddr, *logFile)
	if err := mux.start(); err != nil {
		log.Fatal(err)
	}
}
