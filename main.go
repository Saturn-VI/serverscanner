package main

import (
	"fmt"
	"os"
	"net"
)

var DEFAULT_PORT = 25565

func main() {
	fmt.Println("Hello, World!")
	if len(os.Args) < 2 {
		fmt.Println("Please provide an IP address as an argument")
		return
	}
	fmt.Println(os.Args[1])

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", os.Args[1], DEFAULT_PORT))
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()
	fmt.Println("Connected to server:", conn.RemoteAddr())
}
