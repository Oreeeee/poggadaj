package s2c

import (
	"bytes"
	"encoding/binary"
)

type GG_Recv_MSG struct {
	Sender   uint32
	Seq      uint32
	Time     uint32
	MsgClass uint32
	Content  []byte
}

func (p *GG_Recv_MSG) Serialize() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, p.Sender)
	binary.Write(buf, binary.LittleEndian, p.Seq)
	binary.Write(buf, binary.LittleEndian, p.Time)
	binary.Write(buf, binary.LittleEndian, p.MsgClass)
	binary.Write(buf, binary.LittleEndian, p.Content)

	return buf.Bytes()
}
