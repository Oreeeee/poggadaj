package gg60

import (
	"bytes"
	"encoding/binary"
)

type GG_Login60 struct {
	UIN          uint32
	Hash         uint32
	Status       uint32
	Version      uint32
	Unknown1     byte
	LocalIP      uint32
	LocalPort    uint16
	ExternalIP   uint32
	ExternalPort uint16
	ImageSize    uint8
	Unknown2     byte
}

func (p *GG_Login60) Deserialize(data []byte) {
	buf := bytes.NewBuffer(data)

	binary.Read(buf, binary.LittleEndian, &p.UIN)
	binary.Read(buf, binary.LittleEndian, &p.Hash)
	binary.Read(buf, binary.LittleEndian, &p.Status)
	binary.Read(buf, binary.LittleEndian, &p.Version)
	binary.Read(buf, binary.LittleEndian, &p.Unknown1)
	binary.Read(buf, binary.LittleEndian, &p.LocalIP)
	binary.Read(buf, binary.LittleEndian, &p.LocalPort)
	binary.Read(buf, binary.LittleEndian, &p.ExternalIP)
	binary.Read(buf, binary.LittleEndian, &p.ExternalPort)
	binary.Read(buf, binary.LittleEndian, &p.ImageSize)
	binary.Read(buf, binary.LittleEndian, &p.Unknown2)
}
