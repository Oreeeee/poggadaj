// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package protocol

import (
	"fmt"
	"poggadaj-tcp/utils"
	"strconv"
)

type GG_Userlist_Request struct {
	Type    uint8
	Request []byte
}

func (p *GG_Userlist_Request) Deserialize(stream *utils.IOStream) {
	p.Type = stream.ReadU8()
	p.Request = stream.ReadAll()
}

func (p *GG_Userlist_Request) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("Type: 0x%x", p.Type),
		fmt.Sprintf("Request: %s", strconv.Quote(string(p.Request))),
	}
	return s
}
