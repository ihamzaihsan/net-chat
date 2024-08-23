package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	maxConnections = 10
	maxLineBytes   = 4096
	queueSize      = 64
	writeTimeout   = 5 * time.Second
	nameTimeout    = 30 * time.Second
)

type client struct {
	name string
	conn net.Conn
	out  chan string
	done chan struct{}
}

type server struct {
	mu      sync.Mutex
	clients map[*client]struct{}
	history []string
}

func StartTCPServer(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("listen on port %d: %w", port, err)
	}
	defer listener.Close()
	fmt.Printf("NetChat is listening on port %d\n", port)
	s := &server{clients: make(map[*client]struct{})}
	return s.serve(listener)
}

func (s *server) serve(listener net.Listener) error {
	// Reserve a slot before starting a handler, including unnamed connections.
	slots := make(chan struct{}, maxConnections)
	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("accept connection: %w", err)
		}
		select {
		case slots <- struct{}{}:
			go func() {
				defer func() { <-slots }()
				s.handle(conn)
			}()
		default:
			// A small rejection response has a deadline and allocates no handler.
			if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err == nil {
				if _, err := io.WriteString(conn, "Sorry, but we have reached the maximum number of connections\n"); err != nil {
					log.Printf("reject connection: %v", err)
				}
			}
			conn.Close()
		}
	}
}

func (c *client) enqueue(message string) {
	select {
	case c.out <- message:
	default:
		// Disconnect a client that cannot keep up without blocking the chat.
		c.conn.Close()
	}
}

func (c *client) writeLoop() {
	for {
		select {
		case <-c.done:
			return
		case message := <-c.out:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
				c.conn.Close()
				return
			}
			if _, err := io.WriteString(c.conn, message); err != nil {
				log.Printf("write to %s: %v", c.conn.RemoteAddr(), err)
				c.conn.Close()
				return
			}
		}
	}
}

func (s *server) handle(conn net.Conn) {
	c := &client{conn: conn, out: make(chan string, queueSize), done: make(chan struct{})}
	defer func() {
		s.leave(c)
		close(c.done)
		conn.Close()
	}()
	go c.writeLoop()
	if err := conn.SetReadDeadline(time.Now().Add(nameTimeout)); err != nil {
		return
	}
	c.enqueue(welcomeBanner + "[Enter your name]: ")
	// Reuse the scanner for names and messages so buffered input is preserved.
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1024), maxLineBytes+2)
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if !validName(name) {
			c.enqueue("Name must be 1-64 bytes with no control characters or []:.\n[Enter your name]: ")
			continue
		}
		if !s.join(c, name) {
			c.enqueue("Name already taken. Please choose a different name.\n[Enter your name]: ")
			continue
		}
		if err := conn.SetReadDeadline(time.Time{}); err != nil {
			return
		}
		for scanner.Scan() {
			message := scanner.Text()
			if len(message) > maxLineBytes {
				log.Printf("oversized message from %s", conn.RemoteAddr())
				return
			}
			if strings.TrimSpace(message) == "" {
				continue
			}
			s.broadcast(c, message)
		}
		break
	}
	if err := scanner.Err(); err != nil {
		log.Printf("read from %s: %v", conn.RemoteAddr(), err)
	}
}

func (s *server) join(c *client, name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for existing := range s.clients {
		if existing.name == name {
			return false
		}
	}
	// Name reservation, history replay and registration form one atomic action.
	c.name = name
	c.enqueue(strings.Join(s.history, "") + fmt.Sprintf("Connected as %s. Type a message and press Enter.\n", name))
	for existing := range s.clients {
		existing.enqueue(fmt.Sprintf("%s has joined the chat...\n", name))
	}
	s.clients[c] = struct{}{}
	return true
}

func (s *server) broadcast(sender *client, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	formatted := fmt.Sprintf("[%s][%s]: %s\n", timestamp(), sender.name, message)
	s.history = append(s.history, formatted)
	for c := range s.clients {
		c.enqueue(formatted)
	}
}

func (s *server) leave(c *client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, joined := s.clients[c]; !joined {
		return
	}
	delete(s.clients, c)
	for remaining := range s.clients {
		remaining.enqueue(fmt.Sprintf("%s has left the chat...\n", c.name))
	}
}

func timestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
