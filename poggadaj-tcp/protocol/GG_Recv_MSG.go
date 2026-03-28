// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package protocol

import (
	"fmt"
	"poggadaj-tcp/utils"
)

type GG_Recv_MSG struct {
	Sender   uint32
	Seq      uint32
	Time     uint32
	MsgClass uint32
	Content  string
}

func (p *GG_Recv_MSG) Serialize(stream *utils.IOStream) {
	stream.WriteU32(p.Sender)
	stream.WriteU32(p.Seq)
	stream.WriteU32(p.Time)
	stream.WriteU32(p.MsgClass)
	stream.Write(stream.SerializeString(p.Content, false))
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
