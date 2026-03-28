// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package protocol

import (
	"poggadaj-tcp/utils"
)

type GG_Remove_Notify struct {
	UIN  uint32
	Type byte
}

func (p *GG_Remove_Notify) Deserialize(stream *utils.IOStream) {
	p.UIN = stream.ReadU32()
	p.Type = stream.ReadU8()
}
