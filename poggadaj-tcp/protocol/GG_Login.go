// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package protocol

import (
	"fmt"
	"poggadaj-tcp/utils"
)

type GG_Login struct {
	UIN       uint32
	Hash      uint32
	Status    uint32
	Version   uint32
	LocalIP   uint32
	LocalPort uint16
}

func (p *GG_Login) Deserialize(stream *utils.IOStream) {
	p.UIN = stream.ReadU32()
	p.Hash = stream.ReadU32()
	p.Status = stream.ReadU32()
	p.Version = stream.ReadU32()
	p.LocalIP = stream.ReadU32()
	p.LocalPort = stream.ReadU16()
}

func (p *GG_Login) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("UIN: %d", p.UIN),
		fmt.Sprintf("Hash: %d", p.Hash),
		fmt.Sprintf("Status: 0x%x", p.Status),
		fmt.Sprintf("Version: %d", p.Version),
		fmt.Sprintf("LocalIP: %s", utils.LeIntToIPv4(p.LocalIP).String()),
		fmt.Sprintf("LocalPort: %d", p.LocalPort),
	}
	return s
}
