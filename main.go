package main

import (
	"fmt"
	"net"
	"os"
)

const DEFAULT_PORT = 25565

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide an IP address as an argument")
		return
	}
	fmt.Println(os.Args[1])

	fmt.Println(GetServerStatus(os.Args[1], DEFAULT_PORT))
}

func GetServerStatus(address string, port int) (*ServerStatus, error) {
	// Initial server connection
	// brackets are there for IPv6
	conn, err := net.Dial("tcp", fmt.Sprintf("[%s]:%d", os.Args[1], DEFAULT_PORT))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Send handshake and status request packets
	hs := CreateHandshakePacket(os.Args[1], uint16(DEFAULT_PORT), 1).ToBytes()
	_, err = conn.Write(hs)
	if err != nil {
		return nil, err
	}
	sr := CreateStatusRequestPacket().ToBytes()
	_, err = conn.Write(sr)
	if err != nil {
		return nil, err
	}

	// Read the response from the server
	// This is done in a loop to ensure that the entire packet is read
	// tmp is continuously copied into buf until the entire packet is read
	buf := make([]byte, 0, 4096) // Start with a 0-length slice backed by a 4096-byte array
	tmp := make([]byte, 1024)
	var totalPacketLength int

	for {
		n, err := conn.Read(tmp)
		if err != nil {
			fmt.Println("Error reading response from server:", err)
			return nil, err
		}
		buf = append(buf, tmp[:n]...)

		// if packet length not known yet, read from the server
		if totalPacketLength == 0 {
			packetLength, bytesRead, err := ReadVarInt(buf)
			if err != nil {
				return nil, err
			}
			totalPacketLength = packetLength + bytesRead
		}

		if totalPacketLength != 0 && len(buf) >= totalPacketLength {
			break
		}
	}

	// Decode the response
	// extract packet data from the full packet
	data, _, err := ReadPacket(buf)
	if err != nil {
		return nil, err
	}
	// decode the status info from the packet data
	response, err := DecodeServerStatusResponse(data)
	if err != nil {
		return nil, err
	}

	return ProcessJsonResponse(response)
}
