package c2s

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type GG_Userlist_Request struct {
	Type    uint8
	Request []byte
}

func (p *GG_Userlist_Request) Deserialize(data []byte, packetSize uint32) {
	buf := bytes.NewBuffer(data)

	binary.Read(buf, binary.LittleEndian, &p.Type)
	p.Request = make([]byte, packetSize-1) // idk what the one means but it fixes the reading issue
	binary.Read(buf, binary.LittleEndian, &p.Request)
}

func (p *GG_Userlist_Request) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("Type: 0x%x", p.Type),
		fmt.Sprintf("Request: %s", p.Request),
	}
	return s
}
