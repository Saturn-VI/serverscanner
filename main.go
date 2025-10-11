package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/fxamacker/cbor/v2"
)

const DEFAULT_PORT = 25565
const BADGER_DIR = "./badger"

// worker limit can go up to 20000 before it becomes really not worth it
// so 10k is a good balance
const DEFAULT_WORKERS = 10000

var DEBUG_IP = net.IP{5, 161, 74, 148}

func main() {
	// Parse worker count from args or use default
	workerCount := DEFAULT_WORKERS
	if len(os.Args) > 1 {
		if count, err := strconv.Atoi(os.Args[1]); err == nil && count > 0 {
			workerCount = count
		}
	}

	slog.Info(fmt.Sprintf("Starting with %d workers", workerCount))

	db, err := badger.Open(badger.DefaultOptions(BADGER_DIR))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	jobs := make(chan net.IP, 100)
	results := make(chan *ServerStatus, 100)
	errors := make(chan ErrorWithIP, 100)
	var wg sync.WaitGroup

	for _ = range workerCount {
		wg.Add(1)
		go worker(ctx, jobs, results, errors, &wg)
	}

	go func() {
		jobs <- DEBUG_IP
		SendIPsToChannel(jobs, GenerateAllowedRanges(), done)
	}()

	var readWg sync.WaitGroup
	readWg.Add(1)

	go writer(results, errors, db, &readWg)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Printf("\nReceived signal: %s. Shutting down...\n", sig)
		cancel() // Cancel context to stop workers
		close(done)
	}()

	wg.Wait()
	fmt.Println("All workers finished.")
	// can now safely close these channels
	close(results)
	close(errors)
	fmt.Println("Results and errors channels closed.")
	// wait for writer to finish processing everything
	readWg.Wait()
	fmt.Println("Writer has finished.")
}

var OKAY_ERRORS = []string{
	"dial tcp",
	"read tcp",
	"connection reset by peer",
	"malformed VarInt",
	"context canceled",
}

type ErrorWithIP struct {
	IP   net.IP
	Port int
	Err  error
}

// TODO:
// write data using https://github.com/hypermodeinc/badger
func writer(results <-chan *ServerStatus, errors <-chan ErrorWithIP, db *badger.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case result, ok := <-results:
			if !ok {
				results = nil
			} else {
				processResult(result, db)
			}
		case err, ok := <-errors:
			if !ok {
				errors = nil
			} else {
				slog.Error(err.Err.Error(), "IP", err.IP.String(), "Port", err.Port)
			}
		}

		if results == nil && errors == nil {
			break
		}
	}
}

func processResult(result *ServerStatus, db *badger.DB) {
	fmt.Println("Result:", result.Address, "Version:", result.Version.Name)
	fmt.Println(result.Players.Online, "/", result.Players.Max, "players online")

	// key format is "ip:timestamp"
	tcpAddr, ok := result.Address.(*net.TCPAddr)
	if !ok {
		slog.Error("Address is not a TCPAddr", "address", result.Address.String())
		return
	}

	// len + 1 + 8 because IP + ':' + timestamp (8 bytes)
	key := make([]byte, 0, len(tcpAddr.IP)+1+8)
	key = append(key, tcpAddr.IP...)
	key = append(key, ':')

	// timestamp stuff
	ts := make([]byte, 8)
	binary.BigEndian.PutUint64(ts, uint64(result.Time.Unix()))
	key = append(key, ts...)

	bytes, err := cbor.Marshal(result)
	if err != nil {
		slog.Error("Failed to marshal result", "error", err)
		return
	}

	// write tuah
	// set on that thang
	err = db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, bytes)
	})
	if err != nil {
		slog.Error("Failed to write to database", "error", err)
	}
}

func worker(ctx context.Context, jobs <-chan net.IP, results chan<- *ServerStatus, errors chan<- ErrorWithIP, wg *sync.WaitGroup) {
	defer wg.Done()
WorkLoop:
	for {
		select {
		case <-ctx.Done():
			// Context cancelled, exit
			return
		case ip, ok := <-jobs:
			if !ok {
				// Channel closed, exit
				return
			}
			status, err := GetServerStatus(ctx, ip, DEFAULT_PORT)
			if err != nil {
				for _, err_name := range OKAY_ERRORS {
					if strings.HasPrefix(err.Error(), err_name) {
						continue WorkLoop
					}
				}
				// Send error with context cancellation check
				select {
				case errors <- ErrorWithIP{IP: ip, Port: DEFAULT_PORT, Err: err}:
				case <-ctx.Done():
					return
				}
				continue
			}
			// Send result with context cancellation check
			select {
			case results <- status:
			case <-ctx.Done():
				return
			}
		}
	}
}

func GetServerStatus(ctx context.Context, ip net.IP, port int) (*ServerStatus, error) {
	address := ip.String()
	tcpAddr := &net.TCPAddr{IP: ip, Port: port}

	// Check if context was cancelled before starting
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Initial server connection
	// brackets are there for IPv6
	var d net.Dialer
	d.Timeout = time.Second * 1
	conn, err := d.DialContext(ctx, "tcp", fmt.Sprintf("[%s]:%d", address, DEFAULT_PORT))
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
