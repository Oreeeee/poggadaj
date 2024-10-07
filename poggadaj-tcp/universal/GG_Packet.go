package universal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

type GG_Packet struct {
	PacketType uint32
	Length     uint32
	Data       []byte
}

func InitGG_Packet(packetType uint32, data []byte) *GG_Packet {
	return &GG_Packet{
		PacketType: packetType,
		Length:     uint32(len(data)),
		Data:       data,
	}
}

func (p *GG_Packet) Receive(conn net.Conn) {
	recvBuf := make([]byte, 2048)
	_, err := conn.Read(recvBuf)
	if err != nil {
		fmt.Println("Read Error:", err)
		return
	}

	buf := bytes.NewBuffer(recvBuf)

	binary.Read(buf, binary.LittleEndian, &p.PacketType)
	binary.Read(buf, binary.LittleEndian, &p.Length)
	p.Data = make([]byte, p.Length)
	buf.Read(p.Data)
}

func (p *GG_Packet) Send(conn net.Conn) (int, error) {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, p.PacketType)
	binary.Write(buf, binary.LittleEndian, p.Length)
	buf.Write(p.Data)

	return conn.Write(buf.Bytes())
}
