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

// Scan range limits
var MAX_IP = net.IP{255, 255, 255, 255}
var MIN_IP = net.IP{0, 0, 0, 0}

// Absolute IPv4 bounds, used only to guard against overflow/underflow in
// incrementIP/decrementIP. These are intentionally distinct from MIN_IP/MAX_IP,
// which define the scan region rather than the address space limits.
var ABSOLUTE_MAX_IP = net.IP{255, 255, 255, 255}
var ABSOLUTE_MIN_IP = net.IP{0, 0, 0, 0}

func incrementIP(ip net.IP) net.IP {
	if ip.Equal(ABSOLUTE_MAX_IP) {
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
	if ip.Equal(ABSOLUTE_MIN_IP) {
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
	currentStart := MIN_IP;

	for _, exclude := range EXCLUDE_RANGES {
	  	// Skip excludes entirely below our range
	    if bytes.Compare(exclude.end, currentStart) < 0 {
	        continue
	    }
		// Stop once we've moved past the end of our scan region
		if bytes.Compare(currentStart, MAX_IP) > 0 {
			break
		}
		// If there's a gap between currentStart and the start of the exclude range, add it to allowed
		if bytes.Compare(currentStart, exclude.start) < 0 {
			gapEnd := decrementIP(exclude.start)
			// Clamp the gap to the end of our scan region
			if bytes.Compare(gapEnd, MAX_IP) > 0 {
				gapEnd = MAX_IP
			}
			allowed = append(allowed, IPRange{start: currentStart, end: gapEnd})
		}
		// Move currentStart to the end of the exclude range + 1
		if bytes.Compare(incrementIP(exclude.end), currentStart) > 0 {
	        currentStart = incrementIP(exclude.end)
	    }
	}

	// Add whatever remains up to the end of our scan region
	if bytes.Compare(currentStart, MAX_IP) <= 0 {
		allowed = append(allowed, IPRange{start: currentStart, end: MAX_IP})
	}

	return allowed
}

func SendIPsToChannel(ips chan<- net.IP, ranges []IPRange, done <-chan struct{}) {
	defer close(ips)
	counter := 0
	for _, r := range ranges {
		for ip := r.start; bytes.Compare(ip, r.end) <= 0; ip = incrementIP(ip) {
			select {
			case ips <- ip:
				counter++
				fmt.Printf("Sent: %d\r", counter)
			case <-done:
				fmt.Printf("\nStopping IP generation at %d IPs sent\n", counter)
				return
			}
		}
	}
}
