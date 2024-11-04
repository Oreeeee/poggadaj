package s2c

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type GG_Userlist_Reply struct {
	Type    uint8
	Request []byte
}

func (p *GG_Userlist_Reply) Serialize() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, p.Type)
	binary.Write(buf, binary.LittleEndian, p.Request)
	return buf.Bytes()
}

func (p *GG_Userlist_Reply) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("Type: 0x%x", p.Type),
		fmt.Sprintf("Request: %s", p.Request),
	}
	return s
}
