// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"poggadaj-tcp/utils"

	"golang.org/x/text/encoding/charmap"
)

type GG_Packet struct {
	PacketType uint32
	Length     uint32
	Data       []byte
}

func InitEmptyGG_Packet(packetType uint32) *GG_Packet {
	return &GG_Packet{
		PacketType: packetType,
		Length:     0,
		Data:       []byte{},
	}
}

func InitGG_Packet(packetType uint32, packetStruct GG_Packet_Iface) *GG_Packet {
	data := []byte{}
	if packetStruct != nil {
		stream := utils.NewIOStream([]byte{}, binary.LittleEndian, charmap.Windows1250)
		packetStruct.Serialize(stream)

		data = stream.Get()
	}

	return &GG_Packet{
		PacketType: packetType,
		Length:     uint32(len(data)),
		Data:       data,
	}
}

func ReceivePacket(conn net.Conn) (*GG_Packet, error) {
	// Read PacketType + Length
	recvBuf := make([]byte, 8) // PacketType + Length
	_, err := conn.Read(recvBuf)
	if err != nil {
		return nil, err
	}

	p := &GG_Packet{}

	buf := bytes.NewBuffer(recvBuf)
	binary.Read(buf, binary.LittleEndian, &p.PacketType)
	binary.Read(buf, binary.LittleEndian, &p.Length)

	if p.Length > 0xFFFF {
		// Basic "protection" against crashing the server
		// TODO: Check if this is necessary?
		return nil, errors.New("p.Length > 0xFFFF")
	}

	// Read the rest
	p.Data = make([]byte, p.Length)
	received := 0
	for received < int(p.Length) {
		pass, err := conn.Read(p.Data)
		if err != nil {
			return nil, err
		}
		received += pass
	}

	return p, nil
}

func (p *GG_Packet) Send(conn net.Conn) (int, error) {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, p.PacketType)
	binary.Write(buf, binary.LittleEndian, p.Length)
	buf.Write(p.Data)

	return conn.Write(buf.Bytes())
}
