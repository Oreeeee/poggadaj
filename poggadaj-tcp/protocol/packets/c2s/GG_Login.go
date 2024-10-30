package c2s

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"poggadaj-tcp/utils"
)

type GG_Login struct {
	UIN       uint32
	Hash      uint32
	Status    uint32
	Version   uint32
	LocalIP   uint32
	LocalPort uint16
}

func (p *GG_Login) Deserialize(data []byte) {
	buf := bytes.NewBuffer(data)

	binary.Read(buf, binary.LittleEndian, &p.UIN)
	binary.Read(buf, binary.LittleEndian, &p.Hash)
	binary.Read(buf, binary.LittleEndian, &p.Status)
	binary.Read(buf, binary.LittleEndian, &p.Version)
	binary.Read(buf, binary.LittleEndian, &p.LocalIP)
	binary.Read(buf, binary.LittleEndian, &p.LocalPort)
}

func (p *GG_Login) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("UIN: %d", p.UIN),
		fmt.Sprintf("Hash: %d", p.Hash),
		fmt.Sprintf("Status: 0x%x", p.Status),
		fmt.Sprintf("Version: %d", p.Version),
		fmt.Sprintf("LocalIP: %s", utils.LeIntToIPv4(p.LocalIP).String()),
		fmt.Sprintf("LocalPort: %d", p.LocalPort),
	}
	return s
}
