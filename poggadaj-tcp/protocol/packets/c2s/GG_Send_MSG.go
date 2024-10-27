package c2s

import (
	"bytes"
	"encoding/binary"
)

type GG_Send_MSG struct {
	Recipient uint32
	Seq       uint32
	MsgClass  uint32
	Content   []byte
}

func (p *GG_Send_MSG) Deserialize(data []byte, packetSize uint32) {
	buf := bytes.NewBuffer(data)
	msgSize := packetSize - 8

	binary.Read(buf, binary.LittleEndian, &p.Recipient)
	binary.Read(buf, binary.LittleEndian, &p.Seq)
	binary.Read(buf, binary.LittleEndian, &p.MsgClass)
	p.Content = make([]byte, msgSize)
	buf.Read(p.Content)
}
