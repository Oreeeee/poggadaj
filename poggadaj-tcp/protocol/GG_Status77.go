// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package protocol

import (
	"fmt"
	"poggadaj-tcp/utils"
)

type GG_Status77 struct {
	UIN         uint32
	Status      uint8
	RemoteIP    uint32
	RemotePort  uint16
	Version     uint8
	ImageSize   uint8
	Unknown1    uint8
	Unknown2    uint32
	Description []byte
}

func (p *GG_Status77) Serialize(stream *utils.IOStream) {
	stream.WriteU32(p.UIN)
	stream.WriteU8(p.Status)
	stream.WriteU32(p.RemoteIP)
	stream.WriteU16(p.RemotePort)
	stream.WriteU8(p.Version)
	stream.WriteU8(p.ImageSize)
	stream.WriteU8(p.Unknown1)
	stream.WriteU32(p.Unknown2)
	stream.Write(p.Description)
}

func (p *GG_Status77) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("UIN: %d", p.UIN),
		fmt.Sprintf("Status: 0x%x", p.Status),
		fmt.Sprintf("RemoteIP: %s", utils.LeIntToIPv4(p.RemoteIP).String()),
		fmt.Sprintf("RemotePort: %d", p.RemotePort),
		fmt.Sprintf("Version: %d", p.Version),
		fmt.Sprintf("ImageSize: %d", p.ImageSize),
		fmt.Sprintf("Unknown1: 0x%x", p.Unknown1),
		fmt.Sprintf("Unknown2: 0x%x", p.Unknown2),
		fmt.Sprintf("Description: %s", p.Description),
	}
	return s
}
