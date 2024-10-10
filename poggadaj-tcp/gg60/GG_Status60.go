package gg60

import (
	"bytes"
	"encoding/binary"
)

type GG_Status60 struct {
	UIN        uint32
	Status     uint8
	RemoteIP   uint32
	RemotePort uint16
	Version    uint8
	ImageSize  uint8
	Unknown1   uint8
}

func (p *GG_Status60) Serialize() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, p.UIN)
	binary.Write(buf, binary.LittleEndian, p.Status)
	binary.Write(buf, binary.LittleEndian, p.RemoteIP)
	binary.Write(buf, binary.LittleEndian, p.RemotePort)
	binary.Write(buf, binary.LittleEndian, p.Version)
	binary.Write(buf, binary.LittleEndian, p.ImageSize)
	binary.Write(buf, binary.LittleEndian, p.Unknown1)
	return buf.Bytes()
}
