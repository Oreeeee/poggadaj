// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package protocol

import (
	"fmt"
	"poggadaj-tcp/utils"
)

type GG_Pubdir50_Reply struct {
	Type  uint8
	Seq   uint32
	Reply []byte
}

func (p *GG_Pubdir50_Reply) Serialize(stream *utils.IOStream) {
	stream.WriteU8(p.Type)
	stream.WriteU32(p.Seq)
	stream.Write(p.Reply)
}

func (p *GG_Pubdir50_Reply) Deserialize(stream *utils.IOStream) {
	p.Type = stream.ReadU8()
	p.Seq = stream.ReadU32()
	p.Reply = stream.ReadAll()
}

func (p *GG_Pubdir50_Reply) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("Type: 0x%x", p.Type),
		fmt.Sprintf("Seq: %d", p.Seq),
		fmt.Sprintf("Reply: %s", p.Reply),
	}
	return s
}
