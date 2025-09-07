package main

import (
	"encoding/binary"
)

// https://minecraft.wiki/w/Java_Edition_protocol/Server_List_Ping#Handshake
func CreateHandshakePacket(address string, port uint16, nextState int) Packet {
	var data []byte
	data = append(data, CreateVarInt(-1).bytes...) // -1 means that we're pinging to determine what version to use
	data = append(data, []byte(address)...)
	pb := make([]byte, 2)
	binary.LittleEndian.PutUint16(pb, port)
	data = append(data, pb...)
	data = append(data, CreateVarInt(nextState).bytes...) // 1 for status, 2 for login, 3 for transfer
	return Packet{
		id:   CreateVarInt(0x00), // Handshake packet ID
		data: data,
	}
}
