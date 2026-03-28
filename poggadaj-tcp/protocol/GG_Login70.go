// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package protocol

import (
	"fmt"
	"poggadaj-tcp/utils"
)

type GG_Login70 struct {
	UIN          uint32
	HashType     uint8
	Hash         [64]byte
	Status       uint32
	Version      uint32
	Unknown1     byte
	LocalIP      uint32
	LocalPort    uint16
	ExternalIP   uint32
	ExternalPort uint16
	ImageSize    uint8
	Unknown2     byte
}

func (p *GG_Login70) Deserialize(stream *utils.IOStream) {
	p.UIN = stream.ReadU32()
	p.HashType = stream.ReadU8()
	p.Hash = [64]byte(stream.Read(64))
	p.Status = stream.ReadU32()
	p.Version = stream.ReadU32()
	p.Unknown1 = stream.ReadU8()
	p.LocalIP = stream.ReadU32()
	p.LocalPort = stream.ReadU16()
	p.ExternalIP = stream.ReadU32()
	p.ExternalPort = stream.ReadU16()
	p.ImageSize = stream.ReadU8()
	p.Unknown2 = stream.ReadU8()
}

func (p *GG_Login70) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("UIN: %d", p.UIN),
		fmt.Sprintf("HashType: %d", p.HashType),
		fmt.Sprintf("Hash: 0x%x", p.Hash),
		fmt.Sprintf("Status: %d", p.Status),
		fmt.Sprintf("Version: %d", p.Version),
		fmt.Sprintf("Unknown1: 0x%x", p.Unknown1),
		fmt.Sprintf("LocalIP: %s", utils.LeIntToIPv4(p.LocalIP).String()),
		fmt.Sprintf("LocalPort: %d", p.LocalPort),
		fmt.Sprintf("ExternalIP: %s", utils.LeIntToIPv4(p.ExternalIP).String()),
		fmt.Sprintf("ExternalPort: %d", p.ExternalPort),
		fmt.Sprintf("ImageSize: %d", p.ImageSize),
		fmt.Sprintf("Unknown2: 0x%x", p.Unknown2),
	}
	return s
}
