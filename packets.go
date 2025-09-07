package main

import (
	"fmt"
)

// https://minecraft.wiki/w/Java_Edition_protocol/Server_List_Ping#Handshake
func CreateHandshakePacket(address string, port uint16, nextState int) Packet {
	var data []byte
	// Was previously using -1 as a protocol version, but that doesn't work with modern servers
	// so now just using 0 which is the version for 13w41a and 13w41b (the first versions that use Netty)
	// https://minecraft.wiki/w/Minecraft_Wiki:Projects/wiki.vg_merge/Protocol_version_numbers
	data = append(data, CreateVarInt(0).bytes...)
	data = append(data, CreateString(address).bytes...)
	data = append(data, CreateUnsignedShort(port).bytes...)
	data = append(data, CreateVarInt(nextState).bytes...) // 1 for status, 2 for login, 3 for transfer
	return Packet{
		id:   CreateVarInt(0x00), // Handshake packet ID
		data: data,
	}
}

// https://minecraft.wiki/w/Java_Edition_protocol/Server_List_Ping#Status_Response
func CreateStatusRequestPacket() Packet {
	return Packet{
		id:   CreateVarInt(0x00), // Status Request packet ID
		data: []byte{},
	}
}

// https://minecraft.wiki/w/Java_Edition_protocol/Server_List_Ping#Status_Response
func DecodeServerStatusResponse(packet Packet) (string, error) {
	if packet.id.bytes[0] != 0x00 {
		return "", fmt.Errorf("unexpected packet ID: %x", packet.id.bytes[0])
	}
	response, _, err := ReadString(String{bytes: packet.data})
	if err != nil {
		return "", err
	}
	return response, nil
}
