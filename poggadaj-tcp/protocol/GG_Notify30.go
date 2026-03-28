// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package protocol

import (
	"fmt"
	"poggadaj-tcp/utils"
)

type GG_Notify30 struct {
	UINs []uint32
}

func (p *GG_Notify30) Deserialize(stream *utils.IOStream) {
	contactCount := stream.Available() / 4
	p.UINs = make([]uint32, contactCount)
	for i, _ := range p.UINs {
		p.UINs[i] = stream.ReadU32()
	}
}

func (p *GG_Notify30) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("UINs: %d", p.UINs),
	}
	return s
}
