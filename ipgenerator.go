package main

import (
	"bytes"
	"fmt"
	"net"
)

type IPRange struct {
	start net.IP
	end   net.IP
}

// https://en.wikipedia.org/wiki/Reserved_IP_addresses
var EXCLUDE_RANGES = [...]IPRange{
	{start: net.IP{0, 0, 0, 0}, end: net.IP{0, 255, 255, 255}},
	{start: net.IP{10, 0, 0, 0}, end: net.IP{10, 255, 255, 255}},
	{start: net.IP{100, 64, 0, 0}, end: net.IP{100, 127, 255, 255}},
	{start: net.IP{127, 0, 0, 0}, end: net.IP{127, 255, 255, 255}},
	{start: net.IP{169, 254, 0, 0}, end: net.IP{169, 254, 255, 255}},
	{start: net.IP{172, 16, 0, 0}, end: net.IP{172, 31, 255, 255}},
	{start: net.IP{192, 0, 0, 0}, end: net.IP{192, 0, 0, 255}},
	{start: net.IP{192, 0, 2, 0}, end: net.IP{192, 0, 2, 255}},
	{start: net.IP{192, 88, 99, 0}, end: net.IP{192, 88, 99, 255}},
	{start: net.IP{192, 168, 0, 0}, end: net.IP{192, 168, 255, 255}},
	{start: net.IP{198, 18, 0, 0}, end: net.IP{198, 19, 255, 255}},
	{start: net.IP{198, 51, 100, 0}, end: net.IP{198, 51, 100, 255}},
	{start: net.IP{203, 0, 113, 0}, end: net.IP{203, 0, 113, 255}},
	{start: net.IP{224, 0, 0, 0}, end: net.IP{239, 255, 255, 255}},
	{start: net.IP{233, 252, 0, 0}, end: net.IP{239, 255, 255, 255}},
	{start: net.IP{240, 0, 0, 0}, end: net.IP{255, 255, 255, 255}},
}

var MAX_IP = net.IP{255, 255, 255, 255}
var MIN_IP = net.IP{0, 0, 0, 0}

func incrementIP(ip net.IP) net.IP {
	if ip.Equal(MAX_IP) {
		return ip
	}
	newIP := make(net.IP, len(ip))
	copy(newIP, ip)
	for i := len(newIP) - 1; i >= 0; i-- {
		newIP[i]++
		if newIP[i] != 0 {
			break
		}
	}
	return newIP
}

func decrementIP(ip net.IP) net.IP {
	if ip.Equal(MIN_IP) {
		return ip
	}
	newIP := make(net.IP, len(ip))
	copy(newIP, ip)
	for i := len(newIP) - 1; i >= 0; i-- {
		if newIP[i] > 0 {
			newIP[i]--
			break
		} else {
			newIP[i] = 255
		}
	}
	return newIP
}

func GenerateAllowedRanges() []IPRange {
	var allowed []IPRange

	// Start with the full range
	currentStart := net.IP{0, 0, 0, 0}

	for _, exclude := range EXCLUDE_RANGES {
		// If there's a gap between currentStart and the start of the exclude range, add it to allowed
		if bytes.Compare(currentStart, exclude.start) < 0 {
			allowed = append(allowed, IPRange{start: currentStart, end: decrementIP(exclude.start)})
		}
		// Move currentStart to the end of the exclude range + 1
		currentStart = incrementIP(exclude.end)
	}

	// The last exclude range goes to 255.255.255.255, so this just goes to 240.0.0.0
	if bytes.Compare(currentStart, net.IP{240, 0, 0, 0}) <= 0 {
		allowed = append(allowed, IPRange{start: currentStart, end: net.IP{240, 0, 0, 0}})
	}

	return allowed
}

func SendIPsToChannel(ips chan<- net.IP, ranges []IPRange) {
	counter := 0
	for _, r := range ranges {
		for ip := r.start; bytes.Compare(ip, r.end) <= 0; ip = incrementIP(ip) {
			ips <- ip
			counter++
			fmt.Printf("Sent: %d\r", counter)
		}
	}
	close(ips)
}
