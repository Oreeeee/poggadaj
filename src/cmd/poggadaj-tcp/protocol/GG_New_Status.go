// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package protocol

import (
	"codeberg.org/or3e/poggadaj/cmd/poggadaj-tcp/utils"
)

type GG_New_Status struct {
	Status      uint32
	Description []byte
}

func (p *GG_New_Status) Serialize(stream *utils.IOStream) {}

func (p *GG_New_Status) Deserialize(stream *utils.IOStream) {
	p.Status = stream.ReadU32()
	p.Description = stream.ReadAll()
}
