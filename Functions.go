package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func PrintLinuxLogo(connection net.Conn) {

	file, err := os.Open("linuxLogo.txt")

	if err != nil {
		fmt.Printf("Error: Can't open linuxLogo.txt file\n")
		return
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		connection.Write([]byte(scanner.Text() + "\n"))
	}
}

// check wheather this is a valid port or not
func CheckPort(port int) bool {
	if port < 1024 || port > 65535 {
		fmt.Printf("Error: The port %v is unvalid", port)
		os.Exit(0)
	}
	return true
}

// strconv package is not allowed!
//ASCII to Integer
func Atoi(s string) int {
	value := 0
	sign := 1
	for index, ch := range s {
		if ch == '-' && index == 0 {
			sign = -1
			continue
		} else if ch == '+' && index == 0 {
			sign = 1
			continue
		} else if !(ch >= '0' && ch <= '9') {
			return 0
		}
		//Otherwise,updates value by multiplying by 10 and adding the numeric value of the current character (int(ch-'0'))
		value = value*10 + int(ch-'0')
	}
	 return value * sign
}
//Integer to ASCII
func Itoa(num int) string {
	if num == 0 {
		return "0"
	}

	sign := ""
	if num < 0 {
		sign = "-"
		num = -num
	}

	result := ""
	for num > 0 {
		result = string('0'+(num%10)) + result
		num /= 10
	}
	return sign + result
}
