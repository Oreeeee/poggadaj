// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package protocol

import (
	"fmt"
	"poggadaj-shared/statuses"
	"poggadaj-tcp/utils"
)

type GG_Notify_Reply60 struct {
	UIN            uint32
	Status         uint8
	RemoteIP       uint32
	RemotePort     uint16
	Version        uint8
	ImageSize      uint8
	Unknown1       uint8
	DescriptionLen uint8
	Description    []byte
}

func (p *GG_Notify_Reply60) Serialize(stream *utils.IOStream) {
	// Don't serialize if user not online or invisible
	switch p.Status {
	case statuses.GG_STATUS_NOT_AVAIL:
	case statuses.GG_STATUS_INVISIBLE:
		return
	}

	stream.WriteU32(p.UIN)
	stream.WriteU8(p.Status)
	stream.WriteU32(p.RemoteIP)
	stream.WriteU16(p.RemotePort)
	stream.WriteU8(p.Version)
	stream.WriteU8(p.ImageSize)
	stream.WriteU8(p.Unknown1)
	stream.WriteU8(p.DescriptionLen)
	stream.Write(p.Description)
}

func (p *GG_Notify_Reply60) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("UIN: %d", p.UIN),
		fmt.Sprintf("Status: 0x%x", p.Status),
		fmt.Sprintf("RemoteIP: %s", utils.LeIntToIPv4(p.RemoteIP).String()),
		fmt.Sprintf("RemotePort: %d", p.RemotePort),
		fmt.Sprintf("Version: %d", p.Version),
		fmt.Sprintf("ImageSize: %d", p.ImageSize),
		fmt.Sprintf("Unknown1: %d", p.Unknown1),
		fmt.Sprintf("DescriptionLen: %d", p.DescriptionLen),
		fmt.Sprintf("Description: %s", p.Description),
	}
	return s
}
