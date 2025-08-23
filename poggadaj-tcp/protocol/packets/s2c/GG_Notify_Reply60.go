// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2025 Oreeeee

package s2c

import (
	"bytes"
	"encoding/binary"
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

func (p *GG_Notify_Reply60) Serialize() []byte {
	buf := new(bytes.Buffer)

	// Don't serialize if user not online or invisible
	switch p.Status {
	case statuses.GG_STATUS_NOT_AVAIL:
		return make([]byte, 0)
	case statuses.GG_STATUS_INVISIBLE:
		return make([]byte, 0)
	}

	p.DescriptionLen = uint8(len(p.Description))

	binary.Write(buf, binary.LittleEndian, p.UIN)
	binary.Write(buf, binary.LittleEndian, p.Status)
	binary.Write(buf, binary.LittleEndian, p.RemoteIP)
	binary.Write(buf, binary.LittleEndian, p.RemotePort)
	binary.Write(buf, binary.LittleEndian, p.Version)
	binary.Write(buf, binary.LittleEndian, p.ImageSize)
	binary.Write(buf, binary.LittleEndian, p.Unknown1)
	binary.Write(buf, binary.LittleEndian, p.DescriptionLen)
	binary.Write(buf, binary.LittleEndian, p.Description)
	return buf.Bytes()
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
