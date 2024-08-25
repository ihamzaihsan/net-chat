package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"golang.org/x/term"
)

// Enter is mapped to a callback key so the editor can clear submitted text
// without leaving its echoed line in the transcript.
const submitKey = byte(29)

type terminalInput struct {
	io.Reader
	afterCR bool
}

func (r *terminalInput) Read(buffer []byte) (int, error) {
	for {
		n, err := r.Reader.Read(buffer)
		written := 0
		for _, key := range buffer[:n] {
			if key == '\n' && r.afterCR {
				r.afterCR = false
				continue
			}
			r.afterCR = key == '\r'
			if key == '\r' || key == '\n' {
				key = submitKey
			}
			buffer[written] = key
			written++
		}
		if written > 0 || err != nil {
			return written, err
		}
	}
}

type terminalIO struct {
	io.Reader
	io.Writer
}

func runInteractive(address string, input, output *os.File) error {
	conn, err := net.DialTimeout("tcp", address, networkTimeout)
	if err != nil {
		return fmt.Errorf("connect to %s: %w; check that the server is running on this address", address, err)
	}
	defer conn.Close()
	fd := int(input.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("configure terminal: %w", err)
	}
	defer func() {
		conn.Close()
		term.Restore(fd, state)
	}()

	editor := term.NewTerminal(terminalIO{&terminalInput{Reader: input}, output}, "")
	resize := func() {
		if width, height, err := term.GetSize(int(output.Fd())); err == nil && width > 0 && height > 0 {
			editor.SetSize(width, height)
		}
	}
	resize()
	done := make(chan struct{})
	defer close(done)
	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				resize()
			case <-done:
				return
			}
		}
	}()

	failed := configureSubmission(editor, conn)

	readDone := make(chan error, 1)
	go func() {
		_, err := editor.ReadLine()
		readDone <- err
	}()
	received := make(chan error, 1)
	go func() { received <- receiveInteractive(conn, editor) }()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer signal.Stop(interrupt)
	select {
	case err := <-received:
		if err != nil {
			return fmt.Errorf("receive messages: %w", err)
		}
	case err := <-readDone:
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("read keyboard: %w", err)
		}
	case err := <-failed:
		return err
	case <-interrupt:
	}
	return nil
}

func configureSubmission(editor *term.Terminal, conn net.Conn) <-chan error {
	failed := make(chan error, 1)
	editor.AutoCompleteCallback = func(line string, pos int, key rune) (string, int, bool) {
		if key != rune(submitKey) {
			return "", 0, false
		}
		if len(line) > maxLineBytes {
			editor.Write([]byte(fmt.Sprintf("Input exceeds %d bytes; shorten it before sending.\n", maxLineBytes)))
			return line, pos, true
		}
		err := conn.SetWriteDeadline(time.Now().Add(networkTimeout))
		if err == nil {
			_, err = io.WriteString(conn, line+"\n")
		}
		if err != nil {
			select {
			case failed <- fmt.Errorf("send input: %w", err):
			default:
			}
		}
		return "", 0, true
	}

	return failed
}

func receiveInteractive(input io.Reader, output io.Writer) error {
	reader := bufio.NewReader(input)
	var pending strings.Builder
	for {
		key, err := reader.ReadByte()
		if err != nil {
			if pending.Len() > 0 {
				if _, writeErr := io.WriteString(output, pending.String()+"\n"); writeErr != nil {
					return writeErr
				}
			}
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		pending.WriteByte(key)
		// Put the name request on its own line so all output and editable
		// input positions are tracked correctly by the terminal editor.
		namePrompt := strings.HasSuffix(pending.String(), "[Enter your name]: ")
		if key == '\n' || namePrompt || pending.Len() >= 64*1024 {
			text := pending.String()
			if key != '\n' {
				text += "\n"
			}
			if _, err := io.WriteString(output, text); err != nil {
				return err
			}
			pending.Reset()
		}
	}
}
