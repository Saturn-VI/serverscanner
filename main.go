package main

import (
	"fmt"
	"net"
	"sync"
)

const DEFAULT_PORT = 25565

func main() {
	jobs := make(chan net.IP, 25)
	results := make(chan *ServerStatus, 25)
	errors := make(chan error, 25)

	var wg sync.WaitGroup
	wg.Add(1)
	go worker(&wg, jobs, results, errors)

	jobs <- net.IP{5, 161, 74, 148}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	for result := range results {
		fmt.Println(result.Players)
	}

	fmt.Println(GenerateAllowedRanges())
}

func worker(wg *sync.WaitGroup, jobs <-chan net.IP, results chan<- *ServerStatus, errors chan<- error) {
	defer wg.Done()
	for ip := range jobs {
		status, err := GetServerStatus(ip, DEFAULT_PORT)
		if err != nil {
			errors <- err
			continue
		}
		results <- status
	}
}

func GetServerStatus(ip net.IP, port int) (*ServerStatus, error) {
	address := ip.String()
	tcpAddr := &net.TCPAddr{IP: ip, Port: port}
	// Initial server connection
	// brackets are there for IPv6
	conn, err := net.Dial("tcp", fmt.Sprintf("[%s]:%d", address, DEFAULT_PORT))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Send handshake and status request packets
	hs := CreateHandshakePacket(address, uint16(DEFAULT_PORT), 1).ToBytes()
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

	return ProcessJsonResponse(response, tcpAddr)
}
