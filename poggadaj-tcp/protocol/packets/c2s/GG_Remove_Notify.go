package c2s

import (
	"bytes"
	"encoding/binary"
)

type GG_Remove_Notify struct {
	UIN  uint32
	Type byte
}

func (p *GG_Remove_Notify) Deserialize(data []byte) {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &p.UIN)
	binary.Read(buf, binary.LittleEndian, &p.Type)
}
