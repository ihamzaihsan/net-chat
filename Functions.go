package main

import (
	"errors"
	"fmt"
	"strings"
)

func parsePort(input string) (int, error) {
	if input == "" {
		return 0, errors.New("port must contain digits only")
	}
	port := 0
	for _, digit := range input {
		if digit < '0' || digit > '9' {
			return 0, errors.New("port must contain digits only")
		}
		port = port*10 + int(digit-'0')
		// Reject before the next multiplication, preventing integer overflow.
		if port > 65535 {
			return 0, errors.New("port must be between 1024 and 65535")
		}
	}
	if port < 1024 {
		return 0, errors.New("port must be between 1024 and 65535")
	}
	return port, nil
}

func validName(name string) bool {
	if name == "" || len(name) > 64 || strings.ContainsAny(name, "[]:") {
		return false
	}
	for _, character := range name {
		if character < 32 || character == 127 {
			return false
		}
	}
	return true
}

func prompt(name string) string {
	return fmt.Sprintf("[%s][%s]: ", timestamp(), name)
}

// Keeping the banner in the binary makes it independent of the working directory.
const welcomeBanner = "Welcome to NetChat!\n" + `         _nnnn_
        dGGGGMMb
       @p~qp~~qMb
       M|@||@) M|
       @,----.JM|
      JS^\__/  qKL
     dZP        qKRb
    dZP          qKKb
   fZP            SMMb
   HZM            MMMM
   FqM            MMMM
 __| ".        |\dS"qML
 |    ` + "`" + `.       | ` + "`" + `' \Zq
_)      \.___.,|     .'
\____   )MMMMMP|   .'
     ` + "`" + `-'       ` + "`" + `--'
`
