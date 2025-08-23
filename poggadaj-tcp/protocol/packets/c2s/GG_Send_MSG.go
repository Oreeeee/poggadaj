// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2025 Oreeeee

package c2s

import (
	"bytes"
	"encoding/binary"
	"fmt"
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

func (p *GG_Send_MSG) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("Recipient: %d", p.Recipient),
		fmt.Sprintf("Seq: %d", p.Seq),
		fmt.Sprintf("MsgClass: 0x%x", p.MsgClass),
		fmt.Sprintf("Content: %x", p.Content),
	}
	return s
}
