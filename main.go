package main
 import (
"fmt"
"os"
 )
 func main(){
	switch len(os.Args){
	case 1:
		StartTCPServer(8989)

	case 2:
		port := Atoi(os.Args[1])
		if  port == 0 {
			fmt.Printf("Error: The port must contain numbers only")
			os.Exit(0)
		}

		if CheckPort(port) {
			StartTCPServer(port)
		}
		
	default:
		fmt.Println("[USAGE]: ./TCPChat $port  ")
		os.Exit(0)
	}
}
 