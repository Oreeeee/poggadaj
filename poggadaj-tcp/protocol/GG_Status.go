// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package protocol

import (
	"fmt"
	"poggadaj-tcp/utils"
)

type GG_Status struct {
	UIN         uint32
	Status      uint32
	Description []byte
}

func (p *GG_Status) Serialize(stream *utils.IOStream) {
	stream.WriteU32(p.UIN)
	stream.WriteU32(p.Status)
	stream.Write(p.Description)
}

func (p *GG_Status) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("UIN: %d", p.UIN),
		fmt.Sprintf("Status: 0x%x", p.Status),
		fmt.Sprintf("Description: %s", p.Description),
	}
	return s
}
