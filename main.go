package main

import (
	"fmt"
	"os"
	"net"
)

const DEFAULT_PORT = 25565

func main() {
	fmt.Println("Hello, World!")
	if len(os.Args) < 2 {
		fmt.Println("Please provide an IP address as an argument")
		return
	}
	fmt.Println(os.Args[1])

	// brackets are there for IPv6
	conn, err := net.Dial("tcp", fmt.Sprintf("[%s]:%d", os.Args[1], DEFAULT_PORT))
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()
	fmt.Println("Connected to server:", conn.RemoteAddr())
	hs := CreateHandshakePacket(os.Args[1], uint16(DEFAULT_PORT), 1).ToBytes()
	sr := CreateStatusRequestPacket().ToBytes()

	fmt.Printf("Handshake Packet (%d bytes): %s\n", len(hs), string(hs))
	fmt.Printf("Status Request Packet (%d bytes): %s\n", len(sr), string(sr))
	n, err := conn.Write(hs)
	if err != nil {
		fmt.Println("Error sending handshake packet:", err)
		return
	}
	fmt.Printf("Sent %d bytes for handshake packet\n", n)
	n, err = conn.Write(sr)
	if err != nil {
		fmt.Println("Error sending status request packet:", err)
		return
	}
	fmt.Printf("Sent %d bytes for status request packet\n", n)
	buf := make([]byte, 8192)
	n, err = conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading response from server:", err)
		return
	}
	fmt.Printf("Response from server (%d bytes): %s\n", n, buf[:n])
	// fmt.Printf("First few bytes: % x\n", buf[:n])
}
