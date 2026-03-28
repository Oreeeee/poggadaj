// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package protocol

import (
	"fmt"
	"poggadaj-tcp/utils"
)

type GG_Login30 struct {
	UIN    uint32
	Hash   uint32
	Status uint32
}

func (p *GG_Login30) Deserialize(stream *utils.IOStream) {
	p.UIN = stream.ReadU32()
	p.Hash = stream.ReadU32()
	p.Status = stream.ReadU32()
}

func (p *GG_Login30) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("UIN: %d", p.UIN),
		fmt.Sprintf("Hash: %d", p.Hash),
		fmt.Sprintf("Status: 0x%x", p.Status),
	}
	return s
}
