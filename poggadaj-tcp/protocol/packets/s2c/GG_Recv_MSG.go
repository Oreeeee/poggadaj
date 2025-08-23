// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2025 Oreeeee

package s2c

import (
	"bytes"
	"encoding/binary"
	"fmt"
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

func (p *GG_Recv_MSG) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("Sender: %d", p.Sender),
		fmt.Sprintf("Seq: %d", p.Seq),
		fmt.Sprintf("Time: %d", p.Time),
		fmt.Sprintf("MsgClass: 0x%x", p.MsgClass),
		fmt.Sprintf("Content: %x", p.Content),
	}
	return s
}
