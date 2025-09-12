package main

import (
	"encoding/binary"
	"errors"
)

const SEGMENT_BITS = 0x7F
const CONTINUE_BIT = 0x80

type VarInt struct {
	bytes []byte
}

type String struct {
	bytes []byte
}

type UnsignedShort struct {
	bytes []byte
}

type Packet struct {
	id   VarInt
	data []byte
}

// ts pmo bro

func newTrue() *bool {
	b := true
	return &b
}

func newFalse() *bool {
	b := false
	return &b
}

// https://minecraft.wiki/w/Java_Edition_protocol/Packets#VarInt_and_VarLong
func CreateVarInt(value int) VarInt {
	var bytes []byte
	for {
		if (value & ^SEGMENT_BITS) == 0 {
			bytes = append(bytes, byte(value))
			break
		}

		bytes = append(bytes, byte((value&SEGMENT_BITS)|CONTINUE_BIT))

		value = int(uint(value) >> 7)
	}
	return VarInt{bytes: bytes}
}

func ReadVarInt(data []byte) (value int, bytesRead int, err error) {
	var position uint

	for i := 0; ; i++ {
		// Check if there are enough bytes to read
		if i >= len(data) {
			return 0, 0, errors.New("malformed VarInt: unexpected end of data")
		}

		currentByte := data[i]
		// Get the lower 7 bits
		segment := int(currentByte & SEGMENT_BITS)
		// Shift the segment to its correct position and add to the value
		value |= segment << position

		// If the continue bit is not set, then this is the last byte
		if (currentByte & CONTINUE_BIT) == 0 {
			bytesRead = i + 1
			return value, bytesRead, nil
		}

		position += 7
		// Check for overflow.
		// this is because the minecraft protocol specifies a maximum of 5 bytes for a VarInt
		if position >= 32 {
			return 0, 0, errors.New("VarInt is too big")
		}
	}
}

// https://minecraft.wiki/w/Java_Edition_protocol/Data_types#Type:String
func CreateString(value string) String {
	var bytes []byte
	bytes = append(bytes, CreateVarInt(len(value)).bytes...)
	bytes = append(bytes, []byte(value)...)
	return String{bytes: bytes}
}

func ReadString(data String) (value string, bytesRead int, err error) {
	// strings are prefixed with a VarInt length
	length, n, err := ReadVarInt(data.bytes)
	if err != nil {
		return "", 0, err
	}

	if length < 0 {
		return "", n, errors.New("string length is negative")
	}
	if n+length > len(data.bytes) {
		return "", n, errors.New("not enough bytes to read the full string")
	}

	value = string(data.bytes[n : n+length])
	bytesRead = n + length
	return value, bytesRead, nil
}

// https://minecraft.wiki/w/Java_Edition_protocol/Data_types#Type:Unsigned_Short
func CreateUnsignedShort(value uint16) UnsignedShort {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, value)
	return UnsignedShort{bytes: bytes}
}

// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Packet_format
func (p Packet) ToBytes() []byte {
	length := CreateVarInt(len(p.id.bytes) + len(p.data))
	return append(length.bytes, append(p.id.bytes, p.data...)...)
}

func ReadPacket(data []byte) (packet Packet, bytesRead int, err error) {
	// Packet structure:
	// - Overall packet length (VarInt)
	// - Packet ID (VarInt)
	// - Packet Data (byte array of length (Overall packet length - length of Packet ID))

	// Read the overall packet length (VarInt)
	// This can't be assumed to be the actual size of the packet because some servers send bad data
	packetLength, n, err := ReadVarInt(data)
	if err != nil {
		return Packet{}, 0, err
	}
	bytesRead += n

	if packetLength < 0 {
		return Packet{}, bytesRead, errors.New("packet length is negative")
	}

	// Read the packet ID (VarInt)
	packetID, n, err := ReadVarInt(data[bytesRead:])
	if err != nil {
		return Packet{}, bytesRead, err
	}
	bytesRead += n

	packetDataLength := packetLength - n
	packetData := data[bytesRead : bytesRead+packetDataLength]
	bytesRead += len(packetData)

	return Packet{
		id:   CreateVarInt(packetID),
		data: packetData,
	}, bytesRead, nil
}
