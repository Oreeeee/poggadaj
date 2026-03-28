// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package protocol

import (
	"fmt"
	"poggadaj-tcp/utils"
)

type GG_Pubdir50_Request struct {
	Type    uint8
	Seq     uint32
	Request []byte
}

func (p *GG_Pubdir50_Request) Deserialize(stream *utils.IOStream) {
	p.Type = stream.ReadU8()
	p.Seq = stream.ReadU32()
	p.Request = stream.ReadAll()
}

func (p *GG_Pubdir50_Request) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("Type: 0x%x", p.Type),
		fmt.Sprintf("Seq: %d", p.Seq),
		fmt.Sprintf("Request: %s", p.Request),
	}
	return s
}
