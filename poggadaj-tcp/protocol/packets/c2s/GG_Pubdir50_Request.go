package c2s

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type GG_Pubdir50_Request struct {
	Type    uint8
	Seq     uint32
	Request []byte
}

func (p *GG_Pubdir50_Request) Deserialize(data []byte) {
	buf := bytes.NewBuffer(data)

	binary.Read(buf, binary.LittleEndian, &p.Type)
	binary.Read(buf, binary.LittleEndian, &p.Seq)
	p.Request = make([]byte, len(data)-5)
	buf.Read(p.Request)
}

func (p *GG_Pubdir50_Request) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("Type: 0x%x", p.Type),
		fmt.Sprintf("Seq: %d", p.Seq),
		fmt.Sprintf("Request: %s", p.Request),
	}
	return s
}
