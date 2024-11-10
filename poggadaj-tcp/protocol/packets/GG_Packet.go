package packets

import (
	"bytes"
	"encoding/binary"
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

func (p *GG_Packet) Receive(conn net.Conn) error {
	// Read PacketType + Length
	recvBuf := make([]byte, 8) // PacketType + Length
	_, err := conn.Read(recvBuf)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(recvBuf)
	binary.Read(buf, binary.LittleEndian, &p.PacketType)
	binary.Read(buf, binary.LittleEndian, &p.Length)

	// Read the rest
	p.Data = make([]byte, p.Length)
	_, err = conn.Read(p.Data)
	if err != nil {
		return err
	}

	return nil
}

func (p *GG_Packet) Send(conn net.Conn) (int, error) {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, p.PacketType)
	binary.Write(buf, binary.LittleEndian, p.Length)
	buf.Write(p.Data)

	return conn.Write(buf.Bytes())
}
