package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

const DEFAULT_PORT = 25565

var DEBUG_IP = net.IP{5, 161, 74, 148}

func main() {
	jobs := make(chan net.IP, 100)
	results := make(chan *ServerStatus, 100)
	errors := make(chan error, 100)
	var wg sync.WaitGroup

	for _ = range 100000 {
		wg.Add(1)
		go worker(jobs, results, errors, &wg)
	}

	go func() {
		jobs <- DEBUG_IP
		SendIPsToChannel(jobs, GenerateAllowedRanges())
		close(jobs)
	}()

	var readWg sync.WaitGroup
	readWg.Add(1)

	// go func() {
	// 	defer readWg.Done()
	// 	for err := range errors {
	// 		if strings.HasPrefix(err.Error(), "dial tcp") ||
	// 			strings.HasPrefix(err.Error(), "read tcp") ||
	// 			strings.HasSuffix(err.Error(), "connection reset by peer") {
	// 			continue
	// 		}
	// 		fmt.Println("Error:", err)
	// 	}
	// }()

	// go func() {
	// 	defer readWg.Done()
	// 	for result := range results {
	// 		fmt.Println("Result:", result.Addr, "Version:", result.Version.Name)
	// 		fmt.Println(result.Players.Online, "/", result.Players.Max, "players online")
	// 	}
	// }()

	go writer(results, errors, &readWg)

	wg.Wait()
	close(results)
	close(errors)
	readWg.Wait()
}

// TODO:
// write data using https://github.com/hypermodeinc/badger
func writer(results <-chan *ServerStatus, errors <-chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case result, ok := <-results:
			if !ok {
				results = nil
			} else {
				fmt.Println("Result:", result.Address, "Version:", result.Version.Name)
				fmt.Println(result.Players.Online, "/", result.Players.Max, "players online")
			}
		case err, ok := <-errors:
			if !ok {
				errors = nil
			} else {
				if strings.HasPrefix(err.Error(), "dial tcp") ||
					strings.HasPrefix(err.Error(), "read tcp") ||
					strings.HasSuffix(err.Error(), "connection reset by peer") {
					continue
				}
				fmt.Println("Error:", err)
			}
		}
		if results == nil && errors == nil {
			break
		}
	}
}

func worker(jobs <-chan net.IP, results chan<- *ServerStatus, errors chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	for ip := range jobs {
		status, err := GetServerStatus(ip, DEFAULT_PORT)
		if err != nil {
			errors <- err
			continue
		}
		fmt.Println("worked")
		results <- status
	}
}

func GetServerStatus(ip net.IP, port int) (*ServerStatus, error) {
	address := ip.String()
	tcpAddr := &net.TCPAddr{IP: ip, Port: port}
	// Initial server connection
	// brackets are there for IPv6
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("[%s]:%d", address, DEFAULT_PORT), time.Second*1)
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
			if err.Error() == "EOF" {
				break
			}
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
