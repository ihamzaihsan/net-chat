package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

const (
	defaultAddress = "127.0.0.1:8989"
	maxLineBytes   = 4096
	networkTimeout = 5 * time.Second
)

func main() {
	address, err := parseAddress(os.Args[1:])
	if err == nil {
		err = run(address, os.Stdin, os.Stdout)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, "\nDisconnected from server.")
}

func parseAddress(args []string) (string, error) {
	if len(args) > 1 {
		return "", errors.New("[USAGE]: netchat-client [host:port]")
	}
	if len(args) == 0 {
		return defaultAddress, nil
	}
	address := args[0]
	host, portText, err := net.SplitHostPort(address)
	if err != nil || host == "" || strings.ContainsAny(host, " \t\r\n") {
		return "", errors.New("address must be host:port (for example, 127.0.0.1:8989 or [::1]:8989)")
	}
	port := 0
	for _, digit := range portText {
		if digit < '0' || digit > '9' {
			return "", errors.New("port must contain digits only")
		}
		port = port*10 + int(digit-'0')
		if port > 65535 {
			return "", errors.New("port must be between 1024 and 65535")
		}
	}
	if port < 1024 {
		return "", errors.New("port must be between 1024 and 65535")
	}
	return address, nil
}

func run(address string, input io.Reader, output io.Writer) error {
	conn, err := net.DialTimeout("tcp", address, networkTimeout)
	if err != nil {
		return fmt.Errorf("connect to %s: %w; check that the server is running on this address", address, err)
	}
	defer conn.Close()

	received := make(chan error, 1)
	go func() {
		// Prompts have no newline, so copy bytes as they arrive.
		_, err := io.Copy(output, conn)
		received <- err
	}()
	sent := make(chan error, 1)
	go func() {
		sent <- sendLines(conn, input)
	}()

	// Return on a server disconnect even if keyboard input is still waiting.
	// Closing the socket stops pending network I/O; main exits the process,
	// which also releases a goroutine blocked on terminal input.
	select {
	case err := <-received:
		if err != nil {
			return fmt.Errorf("receive messages: %w", err)
		}
		return nil
	case err := <-sent:
		if err != nil {
			return err
		}
		// EOF closes only the write side, allowing final replies to be read.
		if err := <-received; err != nil {
			return fmt.Errorf("receive messages: %w", err)
		}
		return nil
	}
}

func sendLines(conn net.Conn, input io.Reader) error {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 1024), maxLineBytes+2)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > maxLineBytes {
			return fmt.Errorf("input line exceeds %d bytes", maxLineBytes)
		}
		if err := conn.SetWriteDeadline(time.Now().Add(networkTimeout)); err != nil {
			return fmt.Errorf("set write deadline: %w", err)
		}
		if _, err := io.WriteString(conn, line+"\n"); err != nil {
			return fmt.Errorf("send input: %w", err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read input (maximum line size %d bytes): %w", maxLineBytes, err)
	}
	// DialTimeout with network "tcp" returns a TCP connection.
	return conn.(*net.TCPConn).CloseWrite()
}
