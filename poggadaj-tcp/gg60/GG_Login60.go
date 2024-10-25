package gg60

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"poggadaj-tcp/utils"
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

func (p *GG_Login60) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("UIN: %d", p.UIN),
		fmt.Sprintf("Hash: %d", p.Hash),
		fmt.Sprintf("Status: 0x%x", p.Status),
		fmt.Sprintf("Version: %d", p.Version),
		fmt.Sprintf("Unknown1: 0x%x", p.Unknown1),
		fmt.Sprintf("LocalIP: %s", utils.LeIntToIPv4(p.LocalIP).String()),
		fmt.Sprintf("LocalPort: %d", p.LocalPort),
		fmt.Sprintf("ExternalIP: %s", utils.LeIntToIPv4(p.ExternalIP).String()),
		fmt.Sprintf("ExternalPort: %d", p.ExternalPort),
		fmt.Sprintf("ImageSize: %d", p.ImageSize),
		fmt.Sprintf("Unknown2: 0x%x", p.Unknown2),
	}
	return s
}
