package universal

import (
	"bytes"
	"encoding/binary"
)

type GG_Header struct {
	PacketType uint32
	Length     uint32
	Data       []byte
}

func InitGG_Header(packetType uint32, data []byte) *GG_Header {
	return &GG_Header{
		PacketType: packetType,
		Length:     uint32(len(data)),
		Data:       data,
	}
}

func (h *GG_Header) Serialize() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, h.PacketType)
	binary.Write(buf, binary.LittleEndian, h.Length)
	buf.Write(h.Data)
	return buf.Bytes()
}
