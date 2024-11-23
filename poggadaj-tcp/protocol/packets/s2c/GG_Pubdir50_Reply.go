package s2c

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type GG_Pubdir50_Reply struct {
	Type  uint8
	Seq   uint32
	Reply []byte
}

func (p *GG_Pubdir50_Reply) Serialize() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, p.Type)
	binary.Write(buf, binary.LittleEndian, p.Seq)
	binary.Write(buf, binary.LittleEndian, p.Reply)
	return buf.Bytes()
}

func (p *GG_Pubdir50_Reply) Deserialize(data []byte) {
	buf := bytes.NewBuffer(data)

	binary.Read(buf, binary.LittleEndian, &p.Type)
	binary.Read(buf, binary.LittleEndian, &p.Seq)
	p.Reply = make([]byte, len(data)-5)
	buf.Read(p.Reply)
}

func (p *GG_Pubdir50_Reply) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("Type: 0x%x", p.Type),
		fmt.Sprintf("Seq: %d", p.Seq),
		fmt.Sprintf("Reply: %s", p.Reply),
	}
	return s
}
