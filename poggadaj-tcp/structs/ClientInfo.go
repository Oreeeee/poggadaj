// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2025 Oreeeee

package structs

import (
	"net"
	uv "poggadaj-tcp/universal"
)

type ClientInfo struct {
	Conn          net.Conn
	UIN           uint32
	Status        uint32
	Authenticated bool
	NotifyList    []uv.GG_NotifyContact
}
