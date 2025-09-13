package main

import (
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
	{start: net.IP{233, 252, 0, 0}, end: net.IP{233, 252, 0, 255}},
	{start: net.IP{240, 0, 0, 0}, end: net.IP{255, 255, 255, 255}},
}
