package main

import (
	"fmt"
	"net"
	"os"
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

	// Read the response from the server
	// This is done in a loop to ensure that the entire packet is read
	buf := make([]byte, 0, 4096) // Start with a 0-length slice backed by a 4096-byte array
	tmp := make([]byte, 1024)
	var totalPacketLength int

	for {
		n, err := conn.Read(tmp)
		if err != nil {
			fmt.Println("Error reading response from server:", err)
			return
		}
		buf = append(buf, tmp[:n]...)

		if totalPacketLength == 0 {
			packetLength, bytesRead, err := ReadVarInt(buf)
			if err == nil {
				totalPacketLength = packetLength + bytesRead
			}
		}

		if totalPacketLength != 0 && len(buf) >= totalPacketLength {
			break
		}
	}

	data, n, err := ReadPacket(buf)
	if err != nil {
		fmt.Println("Error reading packet:", err)
		return
	}
	response, err := DecodeServerStatusResponse(data)
	if err != nil {
		fmt.Println("Error decoding server status response:", err)
		return
	}
	fmt.Printf("Server status response length: %d\n", len(response))
	ProcessJsonResponse(response)
}
