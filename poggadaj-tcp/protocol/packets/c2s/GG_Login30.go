// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2025 Oreeeee

package c2s

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type GG_Login30 struct {
	UIN    uint32
	Hash   uint32
	Status uint32
}

func (p *GG_Login30) Deserialize(data []byte) []byte {
	buf := bytes.NewBuffer(data)

	binary.Read(buf, binary.LittleEndian, &p.UIN)
	binary.Read(buf, binary.LittleEndian, &p.Hash)
	binary.Read(buf, binary.LittleEndian, &p.Status)

	return buf.Bytes()
}

func (p *GG_Login30) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("UIN: %d", p.UIN),
		fmt.Sprintf("Hash: %d", p.Hash),
		fmt.Sprintf("Status: 0x%x", p.Status),
	}
	return s
}
