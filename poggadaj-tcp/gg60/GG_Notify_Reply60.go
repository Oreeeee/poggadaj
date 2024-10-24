package gg60

import (
	"bytes"
	"encoding/binary"
	universal "poggadaj-tcp/universal"
)

type GG_Notify_Reply60 struct {
	UIN            uint32
	Status         uint8
	RemoteIP       uint32
	RemotePort     uint16
	Version        uint8
	ImageSize      uint8
	Unknown1       uint8
	DescriptionLen uint8
	Description    []byte
}

func (p *GG_Notify_Reply60) Serialize() []byte {
	buf := new(bytes.Buffer)

	// Don't serialize if user not online or invisible
	switch p.Status {
	case universal.GG_STATUS_NOT_AVAIL:
		return make([]byte, 0)
	case universal.GG_STATUS_INVISIBLE:
		return make([]byte, 0)
	}

	p.DescriptionLen = uint8(len(p.Description))

	binary.Write(buf, binary.LittleEndian, p.UIN)
	binary.Write(buf, binary.LittleEndian, p.Status)
	binary.Write(buf, binary.LittleEndian, p.RemoteIP)
	binary.Write(buf, binary.LittleEndian, p.RemotePort)
	binary.Write(buf, binary.LittleEndian, p.Version)
	binary.Write(buf, binary.LittleEndian, p.ImageSize)
	binary.Write(buf, binary.LittleEndian, p.Unknown1)
	binary.Write(buf, binary.LittleEndian, p.DescriptionLen)
	binary.Write(buf, binary.LittleEndian, p.Description)
	return buf.Bytes()
}
