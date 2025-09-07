package main

const SEGMENT_BITS = 0x7F
const CONTINUE_BIT = 0x80

type VarInt struct {
	bytes []byte
}

type Packet struct {
	id   VarInt
	data []byte
}

// https://minecraft.wiki/w/Java_Edition_protocol/Packets#VarInt_and_VarLong
func CreateVarInt(value int) VarInt {
	var bytes []byte
	for {
		if (value & ^SEGMENT_BITS) == 0 {
			bytes = append(bytes, byte(value))
			break
		}

		bytes = append(bytes, byte((value & SEGMENT_BITS) | CONTINUE_BIT))

		value = int(uint(value) >> 7)
	}
	return VarInt{bytes: bytes}
}

func (p Packet) ToBytes() []byte {
	length := CreateVarInt(len(p.id.bytes) + len(p.data))
	return append(length.bytes, append(p.id.bytes, p.data...)...)
}
