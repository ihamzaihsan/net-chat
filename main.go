package main

import (
	"fmt"
	"os"
)

func main() {
	port := 8989
	if len(os.Args) > 2 {
		fmt.Fprintln(os.Stderr, "[USAGE]: netchat [port]")
		os.Exit(1)
	}
	if len(os.Args) == 2 {
		var err error
		port, err = parsePort(os.Args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	}
	if err := StartTCPServer(port); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
