// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package protocol

import (
	"fmt"
	"poggadaj-tcp/utils"
)

type GG_Send_MSG struct {
	Recipient uint32
	Seq       uint32
	MsgClass  uint32
	Content   string
}

func (p *GG_Send_MSG) Deserialize(stream *utils.IOStream) {
	p.Recipient = stream.ReadU32()
	p.Seq = stream.ReadU32()
	p.MsgClass = stream.ReadU32()
	p.Content = stream.ReadString(-1)
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
