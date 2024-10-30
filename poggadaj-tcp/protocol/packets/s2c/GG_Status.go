package s2c

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type GG_Status struct {
	UIN         uint32
	Status      uint32
	Description []byte
}

func (p *GG_Status) Serialize() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, p.UIN)
	binary.Write(buf, binary.LittleEndian, p.Status)
	binary.Write(buf, binary.LittleEndian, p.Description)
	return buf.Bytes()
}

func (p *GG_Status) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("UIN: %d", p.UIN),
		fmt.Sprintf("Status: 0x%x", p.Status),
		fmt.Sprintf("Description: %s", p.Description),
	}
	return s
}
